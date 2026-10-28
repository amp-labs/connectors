package grain

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/grain/metadata"
	"github.com/spyzhov/ajson"
)

const (
	objectNameRecordings = "recordings"
	objectNameUploads    = "uploads"
)

// includes lists, per object, the fields Grain only returns when asked for in the "include"
// request parameter, with the include that returns each of them.
// https://developers.grain.com/endpoints.html#recording-include
// https://developers.grain.com/endpoints.html#upload-include
var includes = map[string]map[string]map[string]any{ // nolint:gochecknoglobals
	objectNameRecordings: {
		"highlights":      {"highlights": true},
		"participants":    {"participants": true},
		"ai_action_items": {"ai_action_items": true},
		"ai_summary":      {"ai_summary": true},
		"private_notes":   {"private_notes": true},
		"calendar_event":  {"calendar_event": true},
		"hubspot":         {"hubspot": true},
		"screenshares":    {"screenshares": true},
		// Unlike the other includes, this one is an object and rejects true.
		"ai_template_sections": {"ai_template_sections": map[string]string{"format": "json"}},
	},
	objectNameUploads: {
		// The transcript is part of the attributes but has its own include.
		"attributes": {"attributes": true, "transcript": true},
	},
}

// readRequestBody has no page size: Grain has none and rejects any property it does not define.
type readRequestBody struct {
	Cursor  string         `json:"cursor,omitempty"`
	Include map[string]any `json:"include,omitempty"`
}

// buildReadRequest lists one page of an object. Every list is a POST whose JSON body carries
// the cursor of the next page and the includes for the requested fields.
// https://developers.grain.com/endpoints.html
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(readRequestBody{
		Cursor:  params.NextPage.String(),
		Include: requestedIncludes(params),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set(apiVersionHeader, apiVersion)

	return req, nil
}

func requestedIncludes(params common.ReadParams) map[string]any {
	requested := make(map[string]any)

	for field, include := range includes[params.ObjectName] {
		if params.Fields.Has(field) {
			maps.Copy(requested, include)
		}
	}

	return requested
}

// parseReadResponse reads the records held under the object's name, next to the cursor:
//
//	{"cursor": "...", "recordings": [...]}
//
// No list endpoint filters by update time. Uploads carry `updated_datetime`, so Since and Until
// are applied here, both inclusive; the list order is not documented, so every page is read.
// Recordings can only be filtered by when the meeting started (`start_datetime`), which misses
// later edits, so they are read in full, as are hooks, users, teams and meeting types,
// which carry no update time.
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	filter := readhelper.MakeIdentityFilterFunc(nextCursor)
	if params.ObjectName == objectNameUploads {
		filter = readhelper.MakeTimeFilterFunc(
			readhelper.Unordered,
			readhelper.NewTimeBoundary(),
			"updated_datetime", time.RFC3339,
			nextCursor,
		)
	}

	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc(metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)),
		filter,
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// nextCursor returns the cursor of the next page. It is null on the last page of recordings and
// uploads, and absent from hooks, users, teams and meeting types, which are not paginated.
func nextCursor(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("cursor", "")
}
