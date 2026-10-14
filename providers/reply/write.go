package reply

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"strconv"

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

	recordData := prepareRecordData(params)

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

// prepareRecordData lets a ReadResult round-trip into a WriteRequest for
// contacts. Read returns customFields entries as {"key": "<definition id>",
// "value": ...}, while the write endpoints identify a field by "id" (or
// "name") — "key" is not accepted. Entries carrying only a key are rewritten
// to the id form; anything else passes through untouched.
// https://docs.reply.io/api-reference/contacts/update-a-contact
func prepareRecordData(params common.WriteParams) any {
	if params.ObjectName != objectCustomFieldsOwner {
		return params.RecordData
	}

	record, isMap := params.RecordData.(map[string]any)
	if !isMap {
		return params.RecordData
	}

	// ReadResult.Fields lowercases keys, so a round-tripped payload carries
	// "customfields" while a hand-written one carries "customFields".
	fieldsKey := "customFields"

	if _, found := record[fieldsKey]; !found {
		fieldsKey = "customfields"
	}

	fields, ok := record[fieldsKey].([]any)
	if !ok {
		return params.RecordData
	}

	converted := make([]any, len(fields))

	for index, field := range fields {
		entry, ok := field.(map[string]any)
		if !ok {
			converted[index] = field

			continue
		}

		key, hasKey := entry["key"].(string)
		if !hasKey {
			converted[index] = entry

			continue
		}

		id, err := strconv.Atoi(key)
		if err != nil {
			converted[index] = entry

			continue
		}

		converted[index] = map[string]any{"id": id, "value": entry["value"]}
	}

	// Shallow copy so the caller's payload is not mutated.
	result := make(map[string]any, len(record))
	maps.Copy(result, record)

	delete(result, fieldsKey)
	result["customFields"] = converted

	return result
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
