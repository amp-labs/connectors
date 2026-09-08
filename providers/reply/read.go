package reply

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/reply/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Reply accepts up to 1000 records per page (default 25).
	// https://docs.reply.io/api-reference/sequences/list-all-sequences
	defaultPageSize = 100

	objectSequences = "sequences"
)

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

	// Sequences is the only object with a provider-side time filter, and it
	// scopes by creation time; Until has no provider-side parameter, so the
	// full window is enforced connector-side on the created field during
	// parsing. Contacts declare createdAt/lastModifiedAt in their schema, but
	// the live API always returns null for both and the filter endpoint
	// silently ignores time rules, so no other object can be read
	// incrementally (verified against the live API).
	// https://docs.reply.io/api-reference/sequences/list-all-sequences
	if params.ObjectName == objectSequences && !params.Since.IsZero() {
		url.WithQueryParam("createdAfter", datautils.Time.FormatRFC3339inUTC(params.Since))
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

// makeFilterFunc applies the Since/Until window connector-side for sequences —
// the one object whose records carry a populated timestamp (created). Reply
// has no provider-side Until parameter, and its documentation states no sort
// order, so records are treated as unordered. Every other object has no
// usable timestamp and passes through unfiltered.
func makeFilterFunc(params common.ReadParams, request *http.Request) common.RecordsFilterFunc {
	if params.ObjectName != objectSequences {
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(request))
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		"created", time.RFC3339,
		makeNextRecordsURL(request))
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
