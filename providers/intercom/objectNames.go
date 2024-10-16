package intercom

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/intercom/metadata"
)

// Supported object names can be found under schemas.json.
// NOTE: tickets can be queried only by using non-empty `Since` parameter.
var supportedObjectsByRead = handy.NewSetFromList( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"admins":        "admins",
	"teams":         "teams",
	"ticket_types":  "ticket_types",
	"events":        "events",
	"segments":      "segments",
	"activity_logs": "activity_logs",
	"tickets":       "tickets",
	"conversations": "conversations",
},
	func(key string) string {
		// Other objects are mapped to `data`.
		return "data"
	},
)

var objectNameToURLPath = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"activity_logs": "/admins/activity_logs",
	"collections":   "/help_center/collections",
	"help_centers":  "/help_center/help_centers",
	"news_items":    "/news/news_items",
	"newsfeeds":     "/news/newsfeeds",
}, func(obj string) string {
	return obj
})

var incrementalSearchObjectPagination = handy.NewDefaultMap(map[string]int{ //nolint:gochecknoglobals
	// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
	"conversations": 150, // nolint:gomnd
	// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
	"contacts": 50, // nolint:gomnd
	// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/searchtickets
	"tickets": 150, // nolint:gomnd
}, func(k string) int {
	return DefaultPageSize
})
