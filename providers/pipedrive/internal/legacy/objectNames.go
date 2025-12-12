package legacy

import "github.com/amp-labs/connectors/internal/datautils"

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{}, //nolint:gochecknoglobals
	func(key string) string {
		return key
	},
)

var notesFlagFields = datautils.NewSet("pinned_to_deal_flag", "pinned_to_person_flag", // nolint: gochecknoglobals
	"pinned_to_organization_flag", "pinned_to_lead_flag")

var metadataDiscoveryEndpoints = datautils.Map[string, string]{ // nolint: gochecknoglobals
	"activities":    "activityFields",
	"deals":         "dealFields",
	"products":      "productFields",
	"persons":       "personFields",
	"organizations": "organizationFields",
	"notes":         "noteFields",
}
