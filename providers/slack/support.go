package slack

import (
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

// postMethodObjects contains objects whose Slack API endpoint uses HTTP POST instead of GET.
// Pagination params (limit, cursor) are sent in the JSON request body for these objects.
var objectsReadViaPost = datautils.NewSet( //nolint:gochecknoglobals
	// Ref: https://docs.slack.dev/reference/methods/conversations.listConnectInvites
	"conversations.listConnectInvites",

	// Ref: https://docs.slack.dev/reference/methods/conversations.requestSharedInvite.list
	"conversations.requestSharedInvite",
)
