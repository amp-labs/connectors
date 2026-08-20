package mappings

import "net/http"

func init() {
	for name, object := range map[string]Object{
		"chat.scheduledMessages": {
			// https://docs.slack.dev/reference/methods/chat.scheduledMessages.list
			readListInfo: &ReadListInfo{
				Href:            "chat.scheduledMessages.list",
				Method:          http.MethodPost,
				ResponseField:   "scheduled_messages",
				ResponseIdField: "id",
				TimeFilterField: "",
				SinceQP:         "oldest",
				UntilQP:         "latest",
			},
		},
		"auth.teams": {
			// https://docs.slack.dev/reference/methods/auth.teams.list/
			readListInfo: &ReadListInfo{
				Href:            "auth.teams.list",
				Method:          http.MethodPost,
				ResponseField:   "teams",
				ResponseIdField: "id",
				TimeFilterField: "",
			},
		},
	} {
		objectsBotConnector.add(name, object)
		objectsUserConnector.add(name, object)
	}
}
