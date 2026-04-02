package fourfour

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectSinceField maps each object to the OData $filter field used for incremental reads.
// Field names vary per object: some use "created_at", others use "updated".
// Objects not in this map do not support since/until filtering — the default returns empty string.
var objectSinceField = datautils.NewDefaultMap( //nolint:gochecknoglobals
	map[string]string{
		// Objects that support "created_at" field
		"Insights":     "created_at",
		"Topics":       "created_at",
		"TopicModels":  "created_at",
		"Chats":        "created_at",
		"ChatMessages": "created_at",

		// Objects that support "updated" field
		"Accounts":        "updated",
		"Contacts":        "updated",
		"Cases":           "updated",
		"Calls":           "updated",
		"Owners":          "updated",
		"Opportunities":   "updated",
		"Leads":           "updated",
		"Objects":         "updated",
		"TrackerProjects": "updated",
		"TrackerIssues":   "updated",
		"TrackerComments": "updated",
	},
	func(objectName string) string {
		return "" // no filter support
	},
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"Insights",
		"Topics",
		"TopicModels",
		"Conversations",
		"Fragments",
		"Participants",
		"Accounts",
		"Contacts",
		"Cases",
		"Calls",
		"Owners",
		"Opportunities",
		"Leads",
		"Objects",
		"TrackerProjects",
		"TrackerIssues",
		"TrackerComments",
		"Chats",
		"ChatMessages",
		"Labels",
		"CalendarEvents",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
