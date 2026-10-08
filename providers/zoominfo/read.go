package zoominfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
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

	// linksURIPrefix is a ZoomInfo serialization quirk: on the Data API surface the
	// values inside "links" are strings of the form "uri=/data/v1/...?page[number]=2"
	// rather than bare URLs. The Studio and Copilot surfaces return plain absolute
	// URLs instead, so the prefix is trimmed when present.
	linksURIPrefix = "uri="
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
// is derived from the "links" object.
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		getRecords,
		nextPageFromLinks(effectivePageSize(params)),
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

// nextPageFromLinks returns the next page number (as a string token) taken from the
// JSON:API "links" object, or "" when the read is finished.
//
// links.next is the only trustworthy end-of-results signal ZoomInfo provides: it is
// sent while further pages exist and omitted on the final page. It also keeps reads
// inside ZoomInfo's hard 10,000-record pagination limit, past which the API answers
// PFAPI0006 ("Total record pagination is over max allowed value (10000)") — at the
// boundary page links.next simply stops being sent.
//
// meta.page.total is deliberately NOT consulted. Despite its name and ZoomInfo's own
// documentation, it is not a page count: it reports how many records the page just
// returned holds (a full page[size]=100 page reports 100; a 60-record final page
// reports 60). Comparing it against meta.page.number therefore compares a page
// number to a record count, which either walks off the end of the result set into
// PFAPI0004 ("Page number (page) requested is greater than the available results")
// or, at small page sizes, stops after page[size] pages and silently truncates the
// read.
//
// A page shorter than the requested page[size] also ends the read. That is redundant
// with links.next against today's API and is kept so a missing or over-eager link
// cannot resurrect the overrun.
func nextPageFromLinks(pageSize int) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		records, err := getRecords(node)
		if err != nil {
			return "", err
		}

		if pageSize > 0 && len(records) < pageSize {
			return "", nil
		}

		// Unpaginated endpoints (e.g. lookup/{fieldName}) omit "links" entirely.
		links, err := jsonquery.New(node).ObjectOptional("links")
		if err != nil {
			return "", err
		}

		if links == nil {
			return "", nil
		}

		next, err := jsonquery.New(links).StrWithDefault("next", "")
		if err != nil {
			return "", err
		}

		if next == "" {
			return "", nil
		}

		return pageNumberFromLink(next)
	}
}

// pageNumberFromLink extracts page[number] from a "links" entry. A link we cannot
// read is reported as an error rather than treated as end-of-results, so a change in
// ZoomInfo's link format surfaces instead of silently truncating every read.
func pageNumberFromLink(link string) (string, error) {
	parsed, err := neturl.Parse(strings.TrimPrefix(link, linksURIPrefix))
	if err != nil {
		return "", fmt.Errorf("%w: cannot parse links.next %q: %w", common.ErrNextPageInvalid, link, err)
	}

	number := parsed.Query().Get(pageNumberParam)
	if number == "" {
		return "", fmt.Errorf("%w: links.next %q carries no %s", common.ErrNextPageInvalid, link, pageNumberParam)
	}

	return number, nil
}

// effectivePageSize is the page[size] actually sent for a read, clamped to the
// provider maximum. parseReadResponse needs the same value to recognise a short
// final page, so both sides derive it here rather than duplicating the clamp.
func effectivePageSize(params common.ReadParams) int {
	if params.PageSize > 0 && params.PageSize <= maxPageSize {
		return params.PageSize
	}

	return defaultPageSize
}

// applyPagination sets page[size] (capped at the provider max) and page[number]
// (from the opaque NextPage token) on the request URL.
func applyPagination(url *urlbuilder.URL, params common.ReadParams) {
	url.WithQueryParam(pageSizeParam, strconv.Itoa(effectivePageSize(params)))

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
