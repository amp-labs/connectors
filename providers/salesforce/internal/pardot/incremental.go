package pardot

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	responseAt, found := objectsFilterParam[params.ObjectName]
	if !found {
		// No timestamp field available; return all records unfiltered.
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	// Connector side filtering.
	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		responseAt, time.RFC3339,
		nextPageFunc,
	)
}

func nextPageFunc(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextPageUrl", "")
}

// Used for connector side filtering.
var objectsFilterParam = map[string]string{ // nolint:gochecknoglobals
	// Objects that have "updatedAt" field:
	"campaigns":                  "updatedAt",
	"custom-redirects":           "updatedAt",
	"dynamic-contents":           "updatedAt",
	"email-templates":            "updatedAt",
	"engagement-studio-programs": "updatedAt",
	"files":                      "updatedAt",
	"folder-contents":            "updatedAt",
	"folders":                    "updatedAt",
	"form-fields":                "updatedAt",
	"forms":                      "updatedAt",
	"landing-pages":              "updatedAt",
	"layout-templates":           "updatedAt",
	"lifecycle-stages":           "updatedAt",
	"list-emails":                "updatedAt",
	"list-memberships":           "updatedAt",
	"lists":                      "updatedAt",
	"opportunities":              "updatedAt",
	"prospect-accounts":          "updatedAt",
	"prospects":                  "updatedAt",
	"tags":                       "updatedAt",
	"tracker-domains":            "updatedAt",
	"users":                      "updatedAt",
	"visitor-activities":         "updatedAt",
	"visitors":                   "updatedAt",
	"visits":                     "updatedAt",
	// Objects that don't have "updatedAt" but have "createdAt":
	"form-handler-fields": "createdAt",
	"form-handlers":       "createdAt",
	"lifecycle-histories": "createdAt",
	"tagged-objects":      "createdAt",
	"visitor-page-views":  "createdAt",
}

// Used for provider side filtering.
var incrementalSinceQuery = map[string]string{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-v5.html
	"emails": "sentAtAfterOrEqualTo",
}

// Used for provider side filtering.
var incrementalUntilQuery = map[string]string{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-v5.html
	"emails": "sentAtBeforeOrEqualTo",
}
