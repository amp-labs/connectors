package reply

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/reply/metadata"
	"github.com/spyzhov/ajson"
)

// Objects supporting create (POST on the collection root), update and delete
// (verb on /{id}). The remaining read objects have no plain write endpoints:
// linkedin-accounts connect through dedicated flows and inbox/threads is
// read-only.
//
// Sequence creation requires at least one steps entry — the live API rejects
// a body without it ("At least one step is required") even though neither
// Reply's spec nor its documentation lists steps as required.
// https://docs.reply.io/api-reference
//
//nolint:gochecknoglobals
var writeSupport = []string{
	"contacts",
	"contact-accounts",
	"contact-lists",
	"contact-account-lists",
	"sequences",
	"sequence-folders",
	"tasks",
	"email-templates",
	"email-template-folders",
	"email-accounts",
	"custom-fields",
	"schedules",
}

// Reply mixes update verbs across objects: these three update via PATCH
// (e.g. https://docs.reply.io/api-reference/contacts/update-a-contact), every
// other writable object updates via PUT.
//
//nolint:gochecknoglobals
var patchUpdateObjects = datautils.NewStringSet(
	"contacts",
	"email-accounts",
	"sequences",
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// path already carries the /v3 module prefix from the schema registry.
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		if patchUpdateObjects.Has(params.ObjectName) {
			method = http.MethodPatch
		} else {
			method = http.MethodPut
		}
	}

	recordData, err := c.prepareRecordData(ctx, params)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

// prepareRecordData lets a ReadResult round-trip into a WriteRequest. Read
// returns custom fields in two shapes: the provider's customFields entries
// ({"key": "<definition id>", "value": ...}) and one top-level field per
// definition named by its title. Neither is what every write endpoint accepts,
// so both are rewritten into the entry form the endpoint takes (see
// referencesByKey); anything else passes through untouched.
// https://docs.reply.io/api-reference/contacts/update-a-contact
func (c *Connector) prepareRecordData(ctx context.Context, params common.WriteParams) (any, error) {
	if !usesCustomFields(params.ObjectName) {
		return params.RecordData, nil
	}

	record, isMap := params.RecordData.(map[string]any)
	if !isMap {
		return params.RecordData, nil
	}

	// ReadResult.Fields lowercases keys, so a round-tripped payload carries
	// "customfields" while a hand-written one carries "customFields".
	fieldsKey := customFieldsKey

	if _, found := record[fieldsKey]; !found {
		fieldsKey = strings.ToLower(customFieldsKey)
	}

	byKey := referencesByKey(params)

	entries, hasEntries := record[fieldsKey].([]any)
	if hasEntries {
		entries = normalizedEntries(entries, byKey)
	}

	flattened, err := c.flattenedCustomFields(ctx, params.ObjectName, record)
	if err != nil {
		return nil, err
	}

	if !hasEntries && len(flattened) == 0 {
		return params.RecordData, nil
	}

	// Shallow copy so the caller's payload is not mutated.
	result := make(map[string]any, len(record))
	maps.Copy(result, record)

	delete(result, fieldsKey)

	overridden := make(map[int]bool, len(flattened))

	for _, field := range flattened {
		delete(result, field.key)

		overridden[field.id] = true
	}

	entries = withoutOverridden(entries, overridden)

	for _, field := range flattened {
		entries = append(entries, customFieldEntry(field.id, field.value, byKey))
	}

	result[customFieldsKey] = entries

	return result, nil
}

// referencesByKey reports how the write endpoint wants a custom field named.
// Creating a contact takes {"key": ...}, where Reply silently drops an entry
// that carries "id" or "name" instead. Updating a contact and writing an
// account take {"id": ...} (or "name") and reject "key" with a 400.
// Verified live against POST/PATCH /v3/contacts and POST/PUT /v3/contact-accounts.
func referencesByKey(params common.WriteParams) bool {
	return params.ObjectName == "contacts" && !params.IsUpdate()
}

// customFieldEntry builds one customFields entry referencing a definition by id.
func customFieldEntry(id int, value any, byKey bool) map[string]any {
	if byKey {
		return map[string]any{"key": strconv.Itoa(id), "value": stringifyValue(value)}
	}

	return map[string]any{"id": id, "value": stringifyValue(value)}
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := recordIDFromBody(body)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

// recordIDFromBody reads the record id, which Reply types per object:
// integers for most (e.g. contacts {"id": 761575530}), GUID strings for
// sequence-folders — matching the id valueType in the schema registry.
func recordIDFromBody(body *ajson.Node) (string, error) {
	numericID, err := jsonquery.New(body).IntegerOptional("id")
	if err == nil && numericID != nil {
		return strconv.FormatInt(*numericID, 10), nil
	}

	stringID, err := jsonquery.New(body).StringOptional("id")
	if err != nil {
		return "", err
	}

	if stringID == nil {
		return "", nil
	}

	return *stringID, nil
}
