package monaco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

	// Since is documented as strictly-after while Until is up-to-and-including,
	// hence the asymmetric pair of operators.
	conditionGreaterThan      = "greater_than"
	conditionLessThanOrEquals = "less_than_or_equals"

	// tagObjectKey is the required query param of GET /v1/tags/ naming which
	// object type's tags to list.
	tagObjectKey = "object"

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
		if params.ObjectName == objectTags {
			endpointURL, err = withTagObjectParam(endpointURL, params.NextPage)
			if err != nil {
				return nil, err
			}
		}

		return newGetRequest(ctx, endpointURL)
	}

	return newListRequest(ctx, endpointURL, params)
}

// withTagObjectParam appends the required `object` query param to the tags
// URL. Tags are read one object type per page: an empty token starts at the
// first type, and each response's next-page token names the type to fetch
// next, so a full read walks all of tagObjectTypes without Read holding state.
func withTagObjectParam(endpointURL string, token common.NextPageToken) (string, error) {
	objType, err := resolveTagObject(token)
	if err != nil {
		return "", err
	}

	// The types are a fixed lowercase enum, so plain concatenation is safe.
	return endpointURL + "?" + tagObjectKey + "=" + objType, nil
}

// resolveTagObject maps a next-page token onto the tag object type to fetch.
func resolveTagObject(token common.NextPageToken) (string, error) {
	if token == "" {
		return tagObjectTypes[0], nil
	}

	objType := token.String()
	if tagObjectIndex(objType) == -1 {
		return "", fmt.Errorf("%w: expected one of %v, got %q", ErrInvalidNextPage, tagObjectTypes, token)
	}

	return objType, nil
}

func tagObjectIndex(objType string) int {
	for index, candidate := range tagObjectTypes {
		if candidate == objType {
			return index
		}
	}

	return -1
}

func (c *Connector) buildReadURL(objectName string) (string, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), objectName)
	if err != nil {
		return "", err
	}

	modulePath := metadata.Schemas.LookupModuleURLPath(c.ProviderContext.Module())

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, modulePath, path)
	if err != nil {
		return "", err
	}

	result := endpointURL.String()

	// urlbuilder normalizes trailing slashes away, but Monaco's routes are
	// slash-exact: GET /v1/tags/ and /v1/users/ are served directly, while
	// /v1/tags and /v1/users answer 307 pointing at the slashed form. Restore
	// the slash rather than depend on the client following a redirect.
	// /v1/sequence-templates is the reverse -- unslashed is the real route --
	// which is why this keys off the recorded path instead of the object kind.
	if strings.HasSuffix(path, "/") && !strings.HasSuffix(result, "/") {
		result += "/"
	}

	return result, nil
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
		filters = append(filters, filterRule(incrementalField, conditionLessThanOrEquals, params.Until))
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
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		// Records are required, not optional: `data` is mandatory in Monaco's
		// response schema for both the paginated and unpaginated envelopes, so
		// its absence is a contract violation. Treating it as optional would
		// report "zero records, read complete" and let a sync end early and
		// silently; an error is louder and safer.
		common.MakeRecordsFunc(recordsKey),
		makeNextPageFunc(params.ObjectName, req),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func makeNextPageFunc(objectName string, req *http.Request) common.NextPageFunc {
	switch {
	case objectName == objectTags:
		return makeTagsNextPage(req)
	case getListObjects.Has(objectName):
		// The whole collection arrives in one response; there is no pagination
		// object to advance.
		return noNextPage
	default:
		return nextPageFromPagination
	}
}

// makeTagsNextPage advances through tagObjectTypes. The tags response carries
// no pagination block, so the page just fetched is identified by the request's
// own `object` query param and the token returned is simply the next type.
func makeTagsNextPage(req *http.Request) common.NextPageFunc {
	return func(*ajson.Node) (string, error) {
		index := tagObjectIndex(req.URL.Query().Get(tagObjectKey))
		if index == -1 || index+1 == len(tagObjectTypes) {
			return "", nil
		}

		return tagObjectTypes[index+1], nil
	}
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
