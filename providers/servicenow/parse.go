package servicenow

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// recordsPath describes where a GET/list response holds its record array, for
// objects whose response isn't the default `{"result": [...]}` envelope.
type recordsPath struct {
	jsonPath string   // key holding the record array; empty means the root node itself
	nested   []string // intermediate keys to descend through first
}

// objectRecordsPath overrides record extraction for objects whose GET response
// deviates from the default `{"result": [...]}` shape. Verified against the
// Zurich REST API reference. Objects not listed here use the default.
//
// ServiceNow's REST surface isn't a single envelope:
//   - `now`/Table API and many scoped APIs: {"result": [...]}      (default)
//   - SCIM:                                  {"Resources": [...]}
//   - Knowledge / a few scoped APIs:         {"result": {"<key>": [...]}}
//   - TMF / Open APIs:                       [...]  (bare top-level array)
//
// Objects whose list response is a single object rather than an array (e.g. the
// WSD Reservation API) are not supported, since they don't represent a record list.
var objectRecordsPath = map[string]recordsPath{ //nolint:gochecknoglobals
	// SCIM — {"Resources": [...]}
	"Users":  {jsonPath: "Resources"},
	"Groups": {jsonPath: "Resources"},
	"users":  {jsonPath: "Resources"},
	"groups": {jsonPath: "Resources"},

	// Nested array — {"result": {"<key>": [...]}}
	"articles":           {jsonPath: "articles", nested: []string{"result"}},     // Knowledge Management API
	"verifyentitlements": {jsonPath: "entitlements", nested: []string{"result"}}, // Verify Entitlements API
	"installbaseitems":   {jsonPath: "items", nested: []string{"result"}},        // Install Base Item API

	// TMF / Open APIs — bare top-level array [...] (empty spec => the root node)
	"agents":               {jsonPath: ""}, // Agent Client Collector API (/agents/list)
	"alarm":                {jsonPath: ""}, // Alarm Management Open API
	"catalog":              {jsonPath: ""}, // Product Catalog Open API
	"individual":           {jsonPath: ""}, // Party Management Open API
	"lead":                 {jsonPath: ""}, // Lead API
	"organization":         {jsonPath: ""}, // Party Management Open API
	"productOffering":      {jsonPath: ""}, // Product Catalog Open API
	"productOrder":         {jsonPath: ""}, // Product Order Open API
	"productorder":         {jsonPath: ""}, // Product Order Open API
	"productSpecification": {jsonPath: ""}, // Product Catalog Open API
	"quote":                {jsonPath: ""}, // Quote Management API
	"serviceOrder":         {jsonPath: ""}, // Service Order Open API
	"serviceTest":          {jsonPath: ""}, // Service Test Management Open API
	"services":             {jsonPath: ""}, // Cloud Services Catalog API
	"stacks":               {jsonPath: ""}, // Cloud Services Catalog API
	"troubleTicket":        {jsonPath: ""}, // Trouble Ticket Open API
}

// recordsFunc returns the record extractor matching the object's response shape.
// Every shape resolves to common.ExtractRecordsFromPath — including the bare root
// array, where an empty key targets the root node itself.
func recordsFunc(objectName string) common.RecordsFunc {
	spec, ok := objectRecordsPath[objectName]
	if !ok {
		return common.ExtractRecordsFromPath("result")
	}

	return common.ExtractRecordsFromPath(spec.jsonPath, spec.nested...)
}

func getNextRecordsURL(resp *common.JSONHTTPResponse, baseURL string) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		next := httpkit.HeaderLink(resp, "next")

		// Some scoped APIs (e.g. Service Catalog) return a relative Link header;
		// resolve it against the instance base URL so the next page is fetchable.
		if next != "" && !strings.HasPrefix(next, "http") {
			next = strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(next, "/")
		}

		return next, nil
	}
}

// offsetPageSize is the page size requested for offset-paginated objects.
const offsetPageSize = 100

// offsetParams names the query parameters an object paginates with. limitKey may
// be empty for APIs that don't accept a page-size parameter (e.g. the Change API,
// which only takes sysparm_offset and returns a server-controlled page size).
type offsetParams struct {
	limitKey  string // optional; when set, the first page requests offsetPageSize
	offsetKey string
}

// offsetPaginatedObjects use offset-style pagination rather than the Link header,
// requiring the client to advance the offset, and signal the end with an empty
// page. The Knowledge API (limit/offset) qualifies.
//
// The Change API is intentionally excluded: it sends no Link header or total-count,
// and at an offset past the end it returns a duplicate of the last record rather
// than an empty page, so offset pagination can't terminate. Its default read
// returns all records in a single page, so it's read without paging.
var offsetPaginatedObjects = map[string]offsetParams{ //nolint: gochecknoglobals
	"articles": {limitKey: "limit", offsetKey: "offset"},
}

func offsetPaginationOf(objectName string) (offsetParams, bool) {
	p, ok := offsetPaginatedObjects[objectName]

	return p, ok
}

// recordsNodes returns a page's record nodes for an object, respecting its
// response shape (default "result", or the nested path in objectRecordsPath),
// without the per-record map conversion that recordsFunc performs.
func recordsNodes(objectName string, node *ajson.Node) ([]*ajson.Node, error) {
	spec, ok := objectRecordsPath[objectName]
	if !ok {
		spec = recordsPath{jsonPath: "result"}
	}

	return jsonquery.New(node, spec.nested...).ArrayOptional(spec.jsonPath)
}

// offsetNextPage advances the offset by the number of records returned and stops
// once a page comes back empty. Advancing by the actual count (rather than a fixed
// limit) avoids gaps or overlaps for APIs that return a server-controlled page
// size or an extra look-ahead record.
func offsetNextPage(objectName string, request *http.Request) common.NextPageFunc {
	params, _ := offsetPaginationOf(objectName)

	return func(node *ajson.Node) (string, error) {
		records, err := recordsNodes(objectName, node)
		if err != nil {
			return "", err
		}

		if len(records) == 0 {
			return "", nil // empty page => end of records
		}

		query := request.URL.Query()
		offset, _ := strconv.Atoi(query.Get(params.offsetKey))
		query.Set(params.offsetKey, strconv.Itoa(offset+len(records)))

		next := *request.URL
		next.RawQuery = query.Encode()

		return next.String(), nil
	}
}

// pageParams names the page-number/per-page query parameters an object paginates
// with (e.g. the Performance Analytics Scorecards API).
type pageParams struct {
	pageKey    string
	perPageKey string
}

// pagePaginatedObjects use 1-based page-number pagination (page + per-page) and
// signal the end with a short or empty page.
var pagePaginatedObjects = map[string]pageParams{ //nolint: gochecknoglobals
	"scorecards": {pageKey: "sysparm_page", perPageKey: "sysparm_per_page"},
}

func pagePaginationOf(objectName string) (pageParams, bool) {
	p, ok := pagePaginatedObjects[objectName]

	return p, ok
}

// pageNextPage advances the page number by one until a page returns fewer records
// than the requested per-page size, which marks the end.
func pageNextPage(objectName string, request *http.Request) common.NextPageFunc {
	params, _ := pagePaginationOf(objectName)

	return func(node *ajson.Node) (string, error) {
		records, err := recordsNodes(objectName, node)
		if err != nil {
			return "", err
		}

		query := request.URL.Query()

		perPage, err := strconv.Atoi(query.Get(params.perPageKey))
		if err != nil || perPage <= 0 {
			perPage = offsetPageSize
		}

		if len(records) < perPage {
			return "", nil // short page => end of records
		}

		page, err := strconv.Atoi(query.Get(params.pageKey))
		if err != nil || page <= 0 {
			page = 1
		}

		query.Set(params.pageKey, strconv.Itoa(page+1))

		next := *request.URL
		next.RawQuery = query.Encode()

		return next.String(), nil
	}
}
