package monaco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/monaco/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Monaco pages are 1-indexed. page_size accepts 1..500 and defaults to 50;
	// we ask for more per round trip but stay inside the cap.
	firstPage       = 1
	defaultPageSize = 100
	maxPageSize     = 500

	pageKey     = "page"
	pageSizeKey = "page_size"
	filtersKey  = "filters"

	// incrementalField is the record timestamp Since/Until are applied to.
	incrementalField = "updated_at"

	conditionGreaterThan = "greater_than"
	conditionLessThan    = "less_than"

	paginationKey = "pagination"
	pageField     = "page"
	totalPagesKey = "total_pages"
)

var ErrInvalidNextPage = errors.New("invalid next page token")

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	endpointURL, err := c.buildReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if getListObjects.Has(params.ObjectName) {
		return newGetRequest(ctx, endpointURL)
	}

	return newListRequest(ctx, endpointURL, params)
}

func (c *Connector) buildReadURL(objectName string) (string, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), objectName)
	if err != nil {
		return "", err
	}

	modulePath := metadata.Schemas.LookupModuleURLPath(c.ProviderContext.Module())

	// buildURL preserves the trailing slash recorded in schemas.json, which is
	// load-bearing on GET /v1/tags/ and /v1/users/ but must stay off
	// /v1/sequence-templates. See url.go.
	return buildURL(c.ProviderInfo().BaseURL, modulePath, path)
}

func newGetRequest(ctx context.Context, endpointURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func newListRequest(ctx context.Context, endpointURL string, params common.ReadParams) (*http.Request, error) {
	page, err := resolvePage(params.NextPage)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(buildListRequestBody(params, page))
	if err != nil {
		return nil, fmt.Errorf("marshal list request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// buildListRequestBody assembles the POST body. `sort` is deliberately omitted
// so each endpoint keeps its documented default ordering, and
// `include_custom_fields` is omitted so it keeps its default of true on the
// three objects that accept it -- custom fields are not in schemas.json, but
// dropping them from the payload would lose data the caller may want.
func buildListRequestBody(params common.ReadParams, page int) map[string]any {
	body := map[string]any{
		pageKey:     page,
		pageSizeKey: clampPageSize(params.PageSize),
	}

	if filters := buildTimeFilters(params); len(filters) != 0 {
		body[filtersKey] = filters
	}

	return body
}

func buildTimeFilters(params common.ReadParams) []map[string]any {
	if !incrementalObjects.Has(params.ObjectName) {
		return nil
	}

	filters := make([]map[string]any, 0, 2) //nolint:mnd

	if !params.Since.IsZero() {
		filters = append(filters, filterRule(incrementalField, conditionGreaterThan, params.Since))
	}

	if !params.Until.IsZero() {
		filters = append(filters, filterRule(incrementalField, conditionLessThan, params.Until))
	}

	return filters
}

func filterRule(field, condition string, value time.Time) map[string]any {
	return map[string]any{
		"field":     field,
		"condition": condition,
		"value":     value.UTC().Format(time.RFC3339),
	}
}

// resolvePage turns a next-page token into the page to request. Monaco
// paginates by page number rather than by cursor, so the token is just the
// number. Shared with Search, which pages the same endpoints.
func resolvePage(token common.NextPageToken) (int, error) {
	if token == "" {
		return firstPage, nil
	}

	page, err := strconv.Atoi(token.String())
	if err != nil {
		return 0, fmt.Errorf("%w: expected a page number, got %q", ErrInvalidNextPage, token)
	}

	if page < firstPage {
		return 0, fmt.Errorf("%w: pages start at %d, got %d", ErrInvalidNextPage, firstPage, page)
	}

	return page, nil
}

// clampPageSize keeps the requested size inside Monaco's 1..500 range.
func clampPageSize(size int) int {
	switch {
	case size <= 0:
		return defaultPageSize
	case size > maxPageSize:
		return maxPageSize
	default:
		return size
	}
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		extractRecords(recordsKey),
		makeNextPageFunc(params.ObjectName),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func extractRecords(recordsKey string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(recordsKey)
	}
}

func makeNextPageFunc(objectName string) common.NextPageFunc {
	if getListObjects.Has(objectName) {
		// The whole collection arrives in one response; there is no pagination
		// object to advance.
		return noNextPage
	}

	return nextPageFromPagination
}

func noNextPage(*ajson.Node) (string, error) {
	return "", nil
}

// nextPageFromPagination advances while the response says more pages exist.
// A missing or malformed pagination block ends the read rather than erroring,
// so an unexpected payload degrades to a single page instead of a failure.
func nextPageFromPagination(node *ajson.Node) (string, error) {
	pagination, err := jsonquery.New(node).ObjectOptional(paginationKey)
	if err != nil || pagination == nil {
		return "", nil //nolint:nilerr
	}

	page, err := jsonquery.New(pagination).IntegerWithDefault(pageField, 0)
	if err != nil {
		return "", nil //nolint:nilerr
	}

	totalPages, err := jsonquery.New(pagination).IntegerWithDefault(totalPagesKey, 0)
	if err != nil {
		return "", nil //nolint:nilerr
	}

	if page < firstPage || page >= totalPages {
		return "", nil
	}

	return strconv.FormatInt(page+1, 10), nil
}
