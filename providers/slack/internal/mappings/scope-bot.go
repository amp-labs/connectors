package mappings

import (
	"net/http"

	"github.com/amp-labs/connectors/common/readhelper"
)

func init() {
	for name, object := range map[string]Object{
		"conversations.listConnectInvites": {
			// "https://docs.slack.dev/reference/methods/conversations.listconnectinvites"
			readListInfo: &ReadListInfo{
				Href:                  "conversations.listConnectInvites",
				Method:                http.MethodPost,
				ResponseField:         "invites",
				NestedResponseIdField: []string{"invite"},
				ResponseIdField:       "id",
				TimeFilterField:       "date_last_updated",
				// Docs example:
				//	"date_last_updated" -> 1622591318 seconds
				//		= 2021-06-01T23:48:38.000Z
				FilterTimestampFormat: readhelper.TimestampFormatUnixSec,
			},
		},
		"conversations.requestSharedInvite": {
			// "https://docs.slack.dev/reference/methods/conversations.requestsharedinvite.list"
			readListInfo: &ReadListInfo{
				Href:            "conversations.requestSharedInvite.list",
				Method:          http.MethodPost,
				ResponseField:   "invite_requests",
				ResponseIdField: "id",
				TimeFilterField: "date_last_updated",
				// Docs example:
				//	"date_last_updated" -> 1722372331 seconds
				//		= 2024-07-30T20:45:31.000Z
				FilterTimestampFormat: readhelper.TimestampFormatUnixSec,
			},
		},
		"files.remote": {
			readListInfo: nil, // Covered by Bot and User scopes.
			readItemInfo: nil, // Covered by Bot and User scopes.
			// "https://docs.slack.dev/reference/methods/files.remote.add"
			writeCreateInfo: &WriteCreateInfo{
				Href:            "files.remote.add",
				ResponseField:   "file",
				ResponseIdField: "id",
			},
			// "https://docs.slack.dev/reference/methods/files.remote.update"
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "files.remote.update",
				RequestIdField:  "file",
				ResponseField:   "file",
				ResponseIdField: "id",
			},
			// "https://docs.slack.dev/reference/methods/files.remote.remove"
			deleteInfo: &DeleteInfo{
				Href:           "files.remote.remove",
				Method:         http.MethodPost,
				RequestIdField: "file",
			},
		},
		"team.externalTeams": {
			// https://docs.slack.dev/reference/methods/team.externalteams.list
			readListInfo: &ReadListInfo{
				Href:            "team.externalTeams.list",
				Method:          http.MethodGet,
				ResponseField:   "organizations",
				ResponseIdField: "team_id",
			},
		},
	} {
		objectsBotConnector.add(name, object)
	}
}
