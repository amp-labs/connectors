package slack

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectResponseField maps each Slack object name to the JSON key that contains the records
// in the API response.
var objectResponseField = datautils.NewDefaultMap( //nolint:gochecknoglobals
	datautils.Map[string, string]{
		"auth.teams":                        "teams",
		"conversations":                     "channels",
		"conversations.listConnectInvites":  "invites",
		"conversations.requestSharedInvite": "invites",
		"files":                             "files",
		"files.remote":                      "files",
		"reactions":                         "items",
		"team.externalTeams":                "external_teams",
		"usergroups":                        "usergroups",
		"users.conversations":               "channels",
		"users":                             "members",
		"chat.scheduledMessages":            "scheduled_messages",
	},
	func(objectName string) string {
		return objectName
	},
)

// objectsWithoutListSuffix contains objects whose Slack API endpoint does NOT end in ".list".
// All other supported objects are called as "<objectName>.list".
var objectsWithoutListSuffix = datautils.NewSet( //nolint:gochecknoglobals
	"conversations.listConnectInvites",
	"users.conversations",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"auth.teams",
		"conversations",
		"conversations.listConnectInvites",  //POST
		"conversations.requestSharedInvite", // POST
		"files",
		"files.remote",
		"reactions",
		"team.externalTeams",
		"usergroups",
		"users.conversations",
		"users",
		"chat.scheduledMessages",
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
