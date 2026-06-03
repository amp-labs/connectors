package hubspot

import "strings"

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

// referencedObjectTypeMap translates the values HubSpot returns in a property's
// `referencedObjectType` (uppercase singular, e.g. "OWNER") into the object
// names this connector exposes (matching KnownObjectTypes values).
//
// Used by ListObjectMetadata when populating common.FieldMetadata.ReferenceTo
// for reference-typed fields.
var referencedObjectTypeMap = map[string]string{ // nolint:gochecknoglobals
	"APPOINTMENT":     "appointments",
	"CALL":            "calls",
	"COMMUNICATION":   "communications",
	"COMPANY":         "companies",
	"CONTACT":         "contacts",
	"COURSE":          "courses",
	"DEAL":            "deals",
	"EMAIL":           "emails",
	"LEAD":            "leads",
	"LINE_ITEM":       "line_items",
	"LISTING":         "listings",
	"MARKETING_EVENT": "marketing_events",
	"MEETING":         "meetings",
	"NOTE":            "notes",
	"OWNER":           "users",
	"POSTAL_MAIL":     "postal_mail",
	"PRODUCT":         "products",
	"QUOTE":           "quotes",
	"SERVICE":         "services",
	"SUBSCRIPTION":    "subscriptions",
	"TASK":            "tasks",
	"TICKET":          "tickets",
}

// resolveReferencedObjectName maps a HubSpot referencedObjectType value to the
// object name this connector uses. For known core CRM types it returns the
// mapped name from referencedObjectTypeMap. For anything else (custom-object
// FQNs like "p123_my_object", future types we don't know yet) it returns the
// lowercased input so downstream consumers still get a usable reference name.
func resolveReferencedObjectName(hsType string) string {
	if name, ok := referencedObjectTypeMap[hsType]; ok {
		return name
	}

	return strings.ToLower(hsType)
}
