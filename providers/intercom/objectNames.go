package intercom

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/intercom/metadata"
)

// Tickets is a special object which cannot be READ using GET.
// Full read and incremental read are done for tickets using POST search.
const ticketsObjectName = "tickets"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
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

var objectNameToURLPath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"activity_logs": "/admins/activity_logs",
	"collections":   "/help_center/collections",
	"help_centers":  "/help_center/help_centers",
	"news_items":    "/news/news_items",
	"newsfeeds":     "/news/newsfeeds",
}, func(obj string) string {
	return obj
})

// nolint:mnd
var incrementalSearchObjectPagination = datautils.NewDefaultMap(map[string]int{ //nolint:gochecknoglobals
	// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
	"conversations": 150,
	// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
	"contacts": 50,
	// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/searchtickets
	ticketsObjectName: 150,
}, func(k string) int {
	return DefaultPageSize
})
