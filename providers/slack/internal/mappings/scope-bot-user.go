package mappings

import (
	"net/http"

	"github.com/amp-labs/connectors/common/readhelper"
)

func init() { // nolint:funlen,maintidx
	for name, object := range map[string]Object{
		"bookmarks": {
			// https://docs.slack.dev/reference/methods/bookmarks.list
			readListInfo: nil, // Requires `channel_id` query param
			readItemInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/bookmarks.add
			writeCreateInfo: &WriteCreateInfo{
				Href:            "bookmarks.add",
				ResponseField:   "bookmark",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/bookmarks.edit
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "bookmarks.edit",
				RequestIdField:  "bookmark_id",
				ResponseField:   "bookmark",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/bookmarks.remove
			deleteInfo: &DeleteInfo{
				Href:           "bookmarks.remove",
				Method:         http.MethodPost,
				RequestIdField: "bookmark_id", // Looks like channel_id is also required.
			},
		},
		"bots": {
			readListInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/bots.info
			readItemInfo: &ReadItemInfo{
				Href:            "bots.info",
				Method:          http.MethodGet,
				RequestIdField:  "bot",
				ResponseField:   "bot",
				ResponseIdField: "id",
			},
		},
		"calls": {
			readListInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/calls.info
			readItemInfo: &ReadItemInfo{
				Href:            "calls.info",
				Method:          http.MethodPost,
				RequestIdField:  "id",
				ResponseField:   "call",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/calls.add
			writeCreateInfo: &WriteCreateInfo{
				Href:            "calls.add",
				ResponseField:   "call",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/calls.update
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "calls.update",
				RequestIdField:  "id",
				ResponseField:   "call",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/calls.end
			deleteInfo: &DeleteInfo{
				Href:           "calls.end",
				Method:         http.MethodPost,
				RequestIdField: "id",
			},
		},
		"canvases": {
			readListInfo: nil, // N/a
			readItemInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/canvases.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "canvases.create",
				ResponseField:   "",
				ResponseIdField: "canvas_id",
			},
			// https://docs.slack.dev/reference/methods/canvases.edit
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "canvases.edit",
				RequestIdField:  "canvas_id",
				ResponseField:   "",
				ResponseIdField: "", // slackLists.update returns only {"ok": true}; no id in the response.
			},
			// https://docs.slack.dev/reference/methods/canvases.delete
			deleteInfo: &DeleteInfo{
				Href:           "canvases.delete",
				Method:         http.MethodPost,
				RequestIdField: "canvas_id",
			},
		},
		"conversations": {
			// https://docs.slack.dev/reference/methods/conversations.list
			readListInfo: &ReadListInfo{
				Href:            "conversations.list",
				Method:          http.MethodGet,
				ResponseField:   "channels",
				ResponseIdField: "id",
				TimeFilterField: "updated",
			},
			// https://docs.slack.dev/reference/methods/conversations.info
			readItemInfo: &ReadItemInfo{
				Href:            "conversations.info",
				Method:          http.MethodGet,
				RequestIdField:  "channel",
				ResponseField:   "channel",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/conversations.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "conversations.create",
				ResponseField:   "channel",
				ResponseIdField: "id",
			},
			writeUpdateInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/conversations.archive
			deleteInfo: &DeleteInfo{
				Href:           "conversations.archive",
				Method:         http.MethodPost,
				RequestIdField: "channel",
			},
		},
		"conversations.canvases": {
			writeCreateInfo: &WriteCreateInfo{
				Href:            "conversations.canvases.create",
				ResponseField:   "",
				ResponseIdField: "canvas_id",
			},
		},
		"files": {
			// https://docs.slack.dev/reference/methods/files.list
			readListInfo: &ReadListInfo{
				Href:                 "files.list",
				Method:               http.MethodGet,
				ResponseField:        "files",
				ResponseIdField:      "id",
				SinceQP:              "ts_from",
				UntilQP:              "ts_to",
				RangeTimestampFormat: readhelper.TimestampFormatUnixSec,
			},
			// https://docs.slack.dev/reference/methods/files.info
			readItemInfo: &ReadItemInfo{
				Href:            "files.info",
				Method:          http.MethodGet,
				RequestIdField:  "file",
				ResponseField:   "file",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/messaging/working-with-files#uploading_files
			writeCreateInfo: nil, // Uploading files is its own process.
			writeUpdateInfo: nil,
			// https://docs.slack.dev/reference/methods/files.delete
			deleteInfo: &DeleteInfo{
				Href:           "files.delete",
				Method:         http.MethodPost,
				RequestIdField: "file",
			},
		},
		"files.remote": {
			// https://docs.slack.dev/reference/methods/files.remote.list
			readListInfo: &ReadListInfo{
				Href:                 "files.remote.list",
				Method:               http.MethodGet,
				ResponseField:        "files",
				ResponseIdField:      "id",
				SinceQP:              "ts_from",
				UntilQP:              "ts_to",
				RangeTimestampFormat: readhelper.TimestampFormatUnixSec,
			},
			// https://docs.slack.dev/reference/methods/files.remote.info
			readItemInfo: &ReadItemInfo{
				Href:            "files.remote.info",
				Method:          http.MethodGet,
				RequestIdField:  "file",
				ResponseField:   "file",
				ResponseIdField: "id",
			},
		},
		"slackLists": {
			readListInfo: nil, // N/a
			readItemInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/slacklists.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "slackLists.create",
				ResponseField:   "",
				ResponseIdField: "list_id",
			},
			// https://docs.slack.dev/reference/methods/slacklists.update
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "slackLists.update",
				RequestIdField:  "id",
				ResponseField:   "",
				ResponseIdField: "", // slackLists.update returns only {"ok": true}; no id in the response.
			},
			deleteInfo: nil, // N/a
		},
		"slackLists.items": {
			// https://docs.slack.dev/reference/methods/slacklists.items.list
			readListInfo: nil, // Requires query param `list_id`.
			// https://docs.slack.dev/reference/methods/slacklists.items.info
			readItemInfo: nil, // Requires query param `list_id`.
			// https://docs.slack.dev/reference/methods/slacklists.items.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "slackLists.items.create",
				ResponseField:   "item",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/slacklists.items.update
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "slackLists.items.update",
				RequestIdField:  "list_id",
				ResponseField:   "",
				ResponseIdField: "", // slackLists.items.update returns only {"ok": true}; no id in the response.
			},
			// https://docs.slack.dev/reference/methods/slacklists.items.delete
			deleteInfo: nil, // Requires query param `list_id`.
		},
		"team": {
			// https://docs.slack.dev/reference/methods/team.info
			readItemInfo: &ReadItemInfo{
				Href:            "team.info",
				Method:          http.MethodGet,
				RequestIdField:  "team",
				ResponseField:   "team",
				ResponseIdField: "id",
			},
		},
		"usergroups": {
			// https://docs.slack.dev/reference/methods/usergroups.list
			readListInfo: &ReadListInfo{
				Href:            "usergroups.list",
				Method:          http.MethodGet,
				ResponseField:   "usergroups",
				ResponseIdField: "id",
				TimeFilterField: "date_update",
			},
			readItemInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/usergroups.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "usergroups.create",
				ResponseField:   "usergroup",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/usergroups.update
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "usergroups.update",
				RequestIdField:  "usergroup",
				ResponseField:   "usergroup",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/usergroups.disable
			deleteInfo: &DeleteInfo{
				Href:           "usergroups.disable",
				Method:         http.MethodPost,
				RequestIdField: "usergroup",
			},
		},
		// Conversations where self (either Bot/User) is a member of.
		"users.conversations": {
			// https://docs.slack.dev/reference/methods/users.conversations
			readListInfo: &ReadListInfo{
				Href:            "users.conversations",
				Method:          http.MethodGet,
				ResponseField:   "channels",
				ResponseIdField: "id",
				TimeFilterField: "", // Can be filtered by `created` but it is too restrictive
				// and would miss records that should be returned for existing clients.
			},
		},
		"users": {
			// https://docs.slack.dev/reference/methods/users.list
			readListInfo: &ReadListInfo{
				Href:            "users.list",
				Method:          http.MethodGet,
				ResponseField:   "members",
				ResponseIdField: "id",
				TimeFilterField: "updated",
			},
			// https://docs.slack.dev/reference/methods/users.info
			readItemInfo: &ReadItemInfo{
				Href:            "users.info",
				Method:          http.MethodGet,
				RequestIdField:  "user",
				ResponseField:   "user",
				ResponseIdField: "id",
			},
		},
	} {
		objectsBotConnector.add(name, object)
		objectsUserConnector.add(name, object)
	}
}
