package zoominfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = 100
	maxPageSize     = 100

	pageNumberParam = "page[number]"
	pageSizeParam   = "page[size]"
)

// jsonAPIRequestBody is the JSON:API envelope POSTed to a search endpoint.
type jsonAPIRequestBody struct {
	Data jsonAPIRequestData `json:"data"`
}

type jsonAPIRequestData struct {
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

// buildReadRequest constructs a read request for one object. The shape depends on
// the object kind:
//   - search: POST {dataAPIPath}/{resource}/search with the caller's criteria + pagination
//   - lookup: GET  {dataAPIPath}/lookup/{fieldName} (single page)
//   - get:    GET  {segments...} (+ pagination when the endpoint supports it)
//
// Enrich objects are not readable (they are match operations); the endpoint
// registry rejects them before reaching here, but we guard defensively.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	switch kindOf(params.ObjectName) {
	case kindSearch:
		return c.buildSearchReadRequest(ctx, params)
	case kindLookup:
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, segLookup, params.ObjectName)
		if err != nil {
			return nil, err
		}

		return c.newJSONAPIGetRequest(ctx, url)
	case kindGet:
		return c.buildGetReadRequest(ctx, params)
	case kindEnrich, kindUnknown:
		fallthrough
	default:
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, params.ObjectName)
	}
}

// buildGetReadRequest builds a GET list read, applying pagination and the
// incremental updated-since query param where the endpoint supports them.
func (c *Connector) buildGetReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	def := getObjects[params.ObjectName]

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, def.segments...)
	if err != nil {
		return nil, err
	}

	if def.paginated {
		applyPagination(url, params)
	}

	return c.newJSONAPIGetRequest(ctx, url)
}

// buildSearchReadRequest POSTs a search endpoint. The request body carries only
// time criteria derived from ReadParams.Since/Until (the freeform Filter is not
// used). When the object has a required date field, Since defaults to the Unix
// epoch so an unfiltered read still satisfies ZoomInfo's "at least one criterion"
// rule and returns all records.
func (c *Connector) buildSearchReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	def := searchObjects[params.ObjectName]

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, params.ObjectName, "search")
	if err != nil {
		return nil, err
	}

	applyPagination(url, params)

	payload, err := json.Marshal(jsonAPIRequestBody{
		Data: jsonAPIRequestData{Type: def.searchType, Attributes: searchCriteria(def, params)},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIMediaType)
	req.Header.Set("Content-Type", jsonAPIMediaType)

	return req, nil
}

// parseReadResponse turns a JSON:API list response into a ReadResult: records live
// under data[], their fields are flattened out of "attributes", and the next page
// is derived from meta.page.
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		getRecords,
		nextPageFromMeta,
		common.MakeMarshaledDataFunc(common.FlattenNestedFields(attributesField)),
		params.Fields,
	)
}

// getRecords returns the JSON:API resource nodes under data[]. ArrayOptional
// tolerates a missing/null data key (returning no records) so single-object or
// empty responses don't error.
func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	return jsonquery.New(node).ArrayOptional("data")
}

// nextPageFromMeta returns the next page number (as a string token) from
// meta.page, or "" when the current page is the last (or meta.page is absent,
// e.g. lookup endpoints that return a single unpaginated page).
func nextPageFromMeta(node *ajson.Node) (string, error) {
	page := jsonquery.New(node, "meta", "page")

	number, err := page.IntegerOptional("number")
	if err != nil || number == nil {
		return "", nil //nolint:nilerr // absent pagination metadata means a single page
	}

	total, err := page.IntegerOptional("total")
	if err != nil || total == nil {
		return "", nil //nolint:nilerr
	}

	if *number >= *total {
		return "", nil
	}

	return strconv.FormatInt(*number+1, 10), nil
}

// applyPagination sets page[size] (capped at the provider max) and page[number]
// (from the opaque NextPage token) on the request URL.
func applyPagination(url *urlbuilder.URL, params common.ReadParams) {
	size := defaultPageSize
	if params.PageSize > 0 && params.PageSize <= maxPageSize {
		size = params.PageSize
	}

	url.WithQueryParam(pageSizeParam, strconv.Itoa(size))

	if params.NextPage != "" {
		url.WithQueryParam(pageNumberParam, params.NextPage.String())
	}
}

// searchCriteria builds a search object's request attributes from ReadParams,
// using only Since/Until (never Filter):
//   - sinceField is set from Since, defaulting to the Unix epoch (1970-01-01) when
//     Since is unset, so the required date criterion is always present.
//   - untilField is set from Until only when Until is provided.
//
// Objects without a sinceField (e.g. companies) get empty criteria, which their
// search API accepts.
func searchCriteria(def searchDef, params common.ReadParams) map[string]any {
	criteria := map[string]any{}

	if def.sinceField != "" {
		since := params.Since
		if since.IsZero() {
			since = time.Unix(0, 0)
		}

		criteria[def.sinceField] = since.UTC().Format(time.RFC3339)
	}

	if def.untilField != "" && !params.Until.IsZero() {
		criteria[def.untilField] = params.Until.UTC().Format(time.RFC3339)
	}

	return criteria
}
