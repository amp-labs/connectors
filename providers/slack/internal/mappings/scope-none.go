package mappings

import (
	"net/http"

	"github.com/amp-labs/connectors/common/readhelper"
)

func init() {
	for name, object := range map[string]Object{
		"chat.scheduledMessages": {
			// https://docs.slack.dev/reference/methods/chat.scheduledMessages.list
			readListInfo: &ReadListInfo{
				Href:            "chat.scheduledMessages.list",
				Method:          http.MethodPost,
				ResponseField:   "scheduled_messages",
				ResponseIdField: "id",
				TimeFilterField: "date_created",
				// Docs example:
				//	"date_created" -> 1551891734 seconds
				//		= 2019-03-06T17:02:14.000Z
				FilterTimestampFormat: readhelper.TimestampFormatUnixSec,
			},
		},
		"auth.teams": {
			// https://docs.slack.dev/reference/methods/auth.teams.list/
			readListInfo: &ReadListInfo{
				Href:            "auth.teams.list",
				Method:          http.MethodPost,
				ResponseField:   "teams",
				ResponseIdField: "id",
			},
		},
	} {
		objectsBotConnector.add(name, object)
		objectsUserConnector.add(name, object)
	}
}
