package servicenow

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
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
