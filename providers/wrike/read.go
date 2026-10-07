package wrike

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/wrike/metadata"
	"github.com/spyzhov/ajson"
)

// dateRangeFormat is the only timestamp layout Wrike's range filters accept
// (seconds, UTC, no fraction; verified live, millisecond input is rejected).
const dateRangeFormat = "2006-01-02T15:04:05Z"

// dateRange is the JSON value of Wrike's range filters (updatedDate, eventDate).
type dateRange struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	spec, ok := readSpecs[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Subsequent pages are fetched from the exact URL handed back as NextPage.
	// It carries the page token together with the original filters and the
	// `fields` selection, which Wrike does not remember across pages.
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := c.buildFirstPageURL(params, spec)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildFirstPageURL(params common.ReadParams, spec readSpec) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if spec.tokenParam != "" {
		url.WithQueryParam(paramPageSize, strconv.Itoa(pageSize(params, spec)))
	}

	if spec.dateRangeParam != "" {
		if filter, ok := dateRangeFilter(params); ok {
			url.WithQueryParam(spec.dateRangeParam, filter)
		}
	}

	if fields := optionalFieldsRequested(params, spec); fields != "" {
		url.WithQueryParam(paramFields, fields)
	}

	return url, nil
}

func pageSize(params common.ReadParams, spec readSpec) int {
	return min(readhelper.PageSizeWithDefault(params, defaultPageSize), spec.pageLimit)
}

// dateRangeFilter encodes Since/Until as Wrike's {"start","end"} range. Both
// bounds are optional; the whole filter is skipped when neither is set or
// when the range would be empty (Wrike rejects start >= end).
func dateRangeFilter(params common.ReadParams) (string, bool) {
	if params.Since.IsZero() && params.Until.IsZero() {
		return "", false
	}

	if !params.Since.IsZero() && !params.Until.IsZero() && !params.Until.After(params.Since) {
		return "", false
	}

	window := dateRange{}

	if !params.Since.IsZero() {
		window.Start = params.Since.UTC().Format(dateRangeFormat)
	}

	if !params.Until.IsZero() {
		window.End = params.Until.UTC().Format(dateRangeFormat)
	}

	encoded, err := json.Marshal(window)
	if err != nil {
		return "", false
	}

	return string(encoded), true
}

// optionalFieldsRequested returns the JSON array of optional fields the caller
// asked for, in the spelling Wrike expects. Fields are matched
// case-insensitively because ReadParams.Fields may be lowercased.
func optionalFieldsRequested(params common.ReadParams, spec readSpec) string {
	if len(spec.optionalFields) == 0 {
		return ""
	}

	requested := make(map[string]bool, len(params.Fields))
	for _, field := range params.Fields.List() {
		requested[strings.ToLower(field)] = true
	}

	selected := make([]string, 0)

	for _, field := range spec.optionalFields {
		if requested[strings.ToLower(field)] {
			selected = append(selected, field)
		}
	}

	if len(selected) == 0 {
		return ""
	}

	slices.Sort(selected)

	encoded, err := json.Marshal(selected)
	if err != nil {
		return ""
	}

	return string(encoded)
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	spec := readSpecs[params.ObjectName]

	requestURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		recordsFromPath(recordsKey),
		nextPageFunc(spec, requestURL),
		readhelper.MakeMarshaledDataFuncWithId(nil, spec.idField),
		params.Fields,
	)
}

// recordsFromPath extracts the `data` array of the Wrike envelope.
func recordsFromPath(recordsKey string) common.NodeRecordsFunc {
	return func(body *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(body).ArrayOptional(recordsKey)
	}
}

// nextPageFunc turns the envelope's nextPageToken into the next page URL:
// the current request URL with the object's token parameter set. Wrike omits
// the token on the last page.
func nextPageFunc(spec readSpec, requestURL *urlbuilder.URL) common.NextPageFunc {
	if spec.tokenParam == "" {
		return func(*ajson.Node) (string, error) { return "", nil }
	}

	return func(body *ajson.Node) (string, error) {
		token, err := jsonquery.New(body).TextWithDefault(responseNextPageToken, "")
		if err != nil || token == "" {
			return "", err
		}

		nextURL := requestURL.Clone()
		nextURL.WithQueryParam(spec.tokenParam, token)

		return nextURL.String(), nil
	}
}
