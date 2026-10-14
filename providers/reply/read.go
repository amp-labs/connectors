package reply

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/reply/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Reply accepts up to 1000 records per page (default 25).
	// https://docs.reply.io/api-reference/sequences/list-all-sequences
	defaultPageSize = 100
)

// incrementalTimeFields maps an object to its modification-time field; every
// other object is a full read.
// https://docs.reply.io/api-reference/contacts/list-all-contacts
//
//nolint:gochecknoglobals
var incrementalTimeFields = map[string]string{
	"contacts": "lastModifiedAt",
}

// recordTimeLayouts: contacts return zone-less UTC timestamps, other objects RFC 3339.
//
//nolint:gochecknoglobals
var recordTimeLayouts = []string{time.RFC3339, "2006-01-02T15:04:05"}

// Objects whose list endpoint wraps records in {"items": [...], "hasMore": bool}
// paginate with top/skip. The remaining objects return a root-level array in a
// single response, which the empty response key in the schema registry conveys.
func isPaginated(objectName string) bool {
	return metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName) != ""
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// path already carries the /v3 module prefix from the schema registry.
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if isPaginated(params.ObjectName) {
		url.WithQueryParam("top", strconv.Itoa(defaultPageSize))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	transform, err := c.makeRecordTransformer(ctx, params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		response,
		makeGetRecords(params.ObjectName),
		makeFilterFunc(params, request),
		readhelper.MakeMarshaledDataFuncWithId(transform, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// makeFilterFunc keeps records without a parsable timestamp, because
// readhelper.MakeTimeFilterFunc fails the whole read on a null one.
func makeFilterFunc(params common.ReadParams, request *http.Request) common.RecordsFilterFunc {
	nextPage := makeNextRecordsURL(request)

	timeField, incremental := incrementalTimeFields[params.ObjectName]
	if !incremental || (params.Since.IsZero() && params.Until.IsZero()) {
		return readhelper.MakeIdentityFilterFunc(nextPage)
	}

	boundary := readhelper.NewTimeBoundary()

	return func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		next, err := nextPage(body)
		if err != nil {
			return nil, "", err
		}

		filtered := make([]*ajson.Node, 0, len(records))

		for _, record := range records {
			timestamp, found := recordTime(record, timeField)
			if !found || boundary.Contains(params, timestamp) {
				filtered = append(filtered, record)
			}
		}

		return filtered, next, nil
	}
}

func recordTime(record *ajson.Node, field string) (time.Time, bool) {
	value, err := jsonquery.New(record).StringOptional(field)
	if err != nil || value == nil {
		return time.Time{}, false
	}

	for _, layout := range recordTimeLayouts {
		if timestamp, err := time.ParseInLocation(layout, *value, time.UTC); err == nil {
			return timestamp, true
		}
	}

	return time.Time{}, false
}

// makeGetRecords locates the records array: under the schema's response key
// for enveloped objects, or the response root for bare-array objects (empty
// response key).
func makeGetRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		responseKey := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)

		return jsonquery.New(node).ArrayOptional(responseKey)
	}
}

// makeNextRecordsURL advances the skip offset while the response reports
// hasMore. Root-level array responses carry no hasMore, so they finish after
// a single page.
func makeNextRecordsURL(request *http.Request) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasMore, err := jsonquery.New(node).BoolWithDefault("hasMore", false)
		if err != nil || !hasMore {
			return "", err
		}

		url, err := urlbuilder.New(request.URL.String())
		if err != nil {
			return "", err
		}

		skip := 0

		if skipStr, ok := url.GetFirstQueryParam("skip"); ok {
			skip, err = strconv.Atoi(skipStr)
			if err != nil {
				return "", err
			}
		}

		url.WithQueryParam("skip", strconv.Itoa(skip+defaultPageSize))

		return url.String(), nil
	}
}

// makeRecordTransformer resolves contact custom fields to human-readable
// names: the record's customFields entries carry the numeric definition id as
// their key, which is matched against the definition titles fetched once per
// page. Raw is left untouched; the resolved fields are added alongside.
func (c *Connector) makeRecordTransformer(
	ctx context.Context, objectName string,
) (common.RecordTransformer, error) {
	if objectName != objectCustomFieldsOwner {
		return nil, nil // nolint:nilnil
	}

	definitions, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	titleByID := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		titleByID[strconv.Itoa(definition.ID)] = definition.Title
	}

	return func(node *ajson.Node) (map[string]any, error) {
		record, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		fields, ok := record["customFields"].([]any)
		if !ok {
			return record, nil
		}

		for _, field := range fields {
			entry, ok := field.(map[string]any)
			if !ok {
				continue
			}

			key, _ := entry["key"].(string)

			if title, found := titleByID[key]; found {
				record[title] = entry["value"]
			}
		}

		return record, nil
	}, nil
}
