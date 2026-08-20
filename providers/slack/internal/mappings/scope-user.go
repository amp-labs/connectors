package mappings

import (
	"net/http"

	"github.com/amp-labs/connectors/common/readhelper"
)

func init() { // nolint:funlen,maintidx
	for name, object := range map[string]Object{
		"admin.apps.activities": {
			// https://docs.slack.dev/reference/methods/admin.apps.activities.list
			readListInfo: &ReadListInfo{
				Href:                 "admin.apps.activities.list",
				Method:               http.MethodPost,
				ResponseField:        "activities",
				ResponseIdField:      "created", // There is no ID field. Creation time is chosen as a primary key.
				SinceQP:              "min_date_created",
				UntilQP:              "max_date_created",
				RangeTimestampFormat: readhelper.TimestampFormatUnixMs,
			},
		},
		"admin.apps.approved": {
			// https://docs.slack.dev/reference/methods/admin.apps.approved.list
			readListInfo: &ReadListInfo{
				Href:                  "admin.apps.approved.list",
				Method:                http.MethodGet,
				ResponseField:         "approved_apps",
				NestedResponseIdField: []string{"apps"},
				ResponseIdField:       "id",
				TimeFilterField:       "date_updated",
			},
		},
		"admin.apps.requests": {
			// https://docs.slack.dev/reference/methods/admin.apps.requests.list
			readListInfo: &ReadListInfo{
				Href:            "admin.apps.requests.list",
				Method:          http.MethodGet,
				ResponseField:   "app_requests",
				ResponseIdField: "id",
				TimeFilterField: "date_created",
			},
		},
		"admin.apps.restricted": {
			// https://docs.slack.dev/reference/methods/admin.apps.restricted.list
			readListInfo: &ReadListInfo{
				Href:                  "admin.apps.restricted.list",
				Method:                http.MethodGet,
				ResponseField:         "restricted_apps",
				NestedResponseIdField: []string{"app"},
				ResponseIdField:       "id",
				TimeFilterField:       "date_updated",
			},
		},
		"admin.barriers": {
			// https://docs.slack.dev/reference/methods/admin.barriers.list
			readListInfo: &ReadListInfo{
				Href:            "admin.barriers.list",
				Method:          http.MethodGet,
				ResponseField:   "barriers",
				ResponseIdField: "id",
				TimeFilterField: "date_update",
			},
			readItemInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/admin.barriers.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "admin.barriers.create",
				ResponseField:   "barrier",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/admin.barriers.update
			writeUpdateInfo: &WriteUpdateInfo{
				Href:            "admin.barriers.update",
				RequestIdField:  "barrier_id",
				ResponseField:   "barrier",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/admin.barriers.delete
			deleteInfo: &DeleteInfo{
				Href:           "admin.barriers.delete",
				Method:         http.MethodPost,
				RequestIdField: "barrier_id",
			},
		},
		"admin.conversations": {
			readListInfo: nil, // N/a
			readItemInfo: nil, // N/a
			// https://docs.slack.dev/reference/methods/admin.conversations.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "admin.conversations.create",
				ResponseField:   "",
				ResponseIdField: "channel_id",
			},
			// https://docs.slack.dev/reference/methods/admin.conversations.delete
			deleteInfo: &DeleteInfo{
				Href:           "admin.conversations.delete",
				Method:         http.MethodPost,
				RequestIdField: "channel_id",
			},
		},
		"admin.teams": {
			// https://docs.slack.dev/reference/methods/admin.teams.list
			readListInfo: &ReadListInfo{
				Href:            "admin.teams.list",
				Method:          http.MethodPost,
				ResponseField:   "teams",
				ResponseIdField: "id",
			},
			// https://docs.slack.dev/reference/methods/admin.teams.create
			writeCreateInfo: &WriteCreateInfo{
				Href:            "admin.teams.create",
				ResponseField:   "",
				ResponseIdField: "team",
			},
		},
		"admin.users": {
			// https://docs.slack.dev/reference/methods/admin.users.list
			readListInfo: &ReadListInfo{
				Href:            "admin.users.list",
				Method:          http.MethodPost,
				ResponseField:   "users",
				ResponseIdField: "id",
				TimeFilterField: "date_created",
			},
		},
		"admin.users.session": {
			// https://docs.slack.dev/reference/methods/admin.users.session.list
			readListInfo: &ReadListInfo{
				Href:            "admin.users.session.list",
				Method:          http.MethodPost,
				ResponseField:   "active_sessions",
				ResponseIdField: "session_id",
			},
		},
		"stars": {
			// https://docs.slack.dev/reference/methods/stars.list
			// End-users can use the new Later view, but Later APIs are not currently available.
			// As a result of this transition, the stars.list method will no longer
			// reflect anything new that users are saving.
			readListInfo: &ReadListInfo{
				Href:                  "stars.list",
				Method:                http.MethodGet,
				ResponseField:         "items",
				NestedResponseIdField: []string{"message"},
				ResponseIdField:       "permalink",
				TimeFilterField:       "date_create",
			},
		},
	} {
		objectsUserConnector.add(name, object)
	}
}
