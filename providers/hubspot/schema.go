package hubspot

// KnownObjectTypes
// https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
var KnownObjectTypes = map[string]string{ // nolint:gochecknoglobals
	"0-2":   "companies",
	"0-1":   "contacts",
	"0-3":   "deals",
	"0-5":   "tickets",
	"0-421": "appointments",
	"0-48":  "calls",
	"0-18":  "communications",
	"0-410": "courses",
	"0-49":  "emails",
	"0-136": "leads",
	"0-8":   "line_items",
	"0-420": "listings",
	"0-54":  "marketing_events",
	"0-47":  "meetings",
	"0-46":  "notes",
	"0-116": "postal_mail",
	"0-7":   "products",
	"0-14":  "quotes",
	"0-162": "services",
	"0-69":  "subscriptions",
	"0-27":  "tasks",
	"0-115": "users",
}
