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

// objectsWithConnectorSideFilter maps each object that supports connector-side time filtering
// to the JSON field used for comparison. Slack has no server-side date filter params,
// so filtering is done in memory after each page is fetched. All Slack timestamps are
// Unix epoch seconds.
var objectsWithConnectorSideFilter = datautils.Map[string, string]{ //nolint:gochecknoglobals
	"conversations":       "updated",
	"files":               "created",
	"usergroups":          "date_update",
	"users.conversations": "created",
	"users":               "updated",
}

// objectsUsingAddSuffix contains write objects whose Slack API endpoint uses the ".add" method suffix.
// All other supported write objects use ".create".
var writeObjectsUsingAddSuffix = datautils.NewSet( //nolint:gochecknoglobals
	"calls",
	"bookmarks",
	"files.remote",
)

// writeResponseSpec describes how to extract the created record from a Slack write response.
type writeResponseSpec struct {
	// recordKey is the key for the created object (e.g. "channel", "reminder").
	// Empty means the ID lives at the response root.
	recordKey string
	// idField is the field that holds the record ID. Empty means this object returns no ID.
	idField string
}

// writeResponseField maps each supported write object (base name, without suffix) to
// the shape of its API response, used to extract the record ID and data after a successful write.
//
//nolint:gochecknoglobals
var writeResponseField = datautils.Map[string, writeResponseSpec]{
	"calls":                  {"call", "id"},
	"bookmarks":              {"bookmark", "id"},
	"canvases":               {"", "canvas_id"},
	"conversations.canvases": {"", "canvas_id"},
	"conversations":          {"channel", "id"},
	"files.remote":           {"file", "id"},
	"slackLists":             {"", "list_id"},
	"slackLists.items":       {"item", "id"},
	"usergroups":             {"usergroup", "id"},
}

// writeUpdateSuffix maps each write object that supports updates (base name) to the
// Slack API method suffix used for the update call (e.g. ".update", ".edit").
//
//nolint:gochecknoglobals
var writeUpdateSuffix = datautils.Map[string, string]{
	"calls":        ".update",
	"bookmarks":    ".edit",
	"canvases":     ".edit",
	"files.remote": ".update",
	"slackLists":   ".update",
	"usergroups":   ".update",
}

// writeUpdateIdField maps each updatable write object to the request body field
// used to pass the record ID to the Slack update endpoint. Objects absent from this map do
// not require an ID in the request body.
//
//nolint:gochecknoglobals
var writeUpdateIdField = datautils.Map[string, string]{
	"calls":      "id",
	"bookmarks":  "bookmark_id",
	"canvases":   "canvas_id",
	"slackLists": "id",
	"usergroups": "usergroup",
}
