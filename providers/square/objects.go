package square

const apiVersion = "v2"

// objectConfig describes how to list an object's records on the Square API.
type objectConfig struct {
	// path is the URL path (after the base URL and API version) that lists the
	// object's records, e.g. "/customers" or "/catalog/list".
	path string
	// responseKey is the JSON key in the response that holds the records array.
	responseKey string
	// supportsLimit reports whether the endpoint accepts a `limit` query param.
	supportsLimit bool
	// supportsCursor reports whether the endpoint paginates via a `cursor` query param.
	supportsCursor bool
	// supportsTimeRange reports whether the endpoint filters by creation time via
	// `begin_time`/`end_time` query params, enabling incremental reads.
	supportsTimeRange bool
}

// objects is the set of objects the Square connector supports. Each exposes a
// GET list endpoint that returns an array of records under responseKey.
var objects = map[string]objectConfig{ //nolint:gochecknoglobals
	"customers": {
		path:           "/customers",
		responseKey:    "customers",
		supportsLimit:  true,
		supportsCursor: true,
	},
	"locations": {
		path:        "/locations",
		responseKey: "locations",
	},
	"payments": {
		path:              "/payments",
		responseKey:       "payments",
		supportsLimit:     true,
		supportsCursor:    true,
		supportsTimeRange: true,
	},
	"refunds": {
		path:              "/refunds",
		responseKey:       "refunds",
		supportsLimit:     true,
		supportsCursor:    true,
		supportsTimeRange: true,
	},
	"catalog": {
		path:           "/catalog/list",
		responseKey:    "objects",
		supportsCursor: true,
	},
	"cards": {
		path:           "/cards",
		responseKey:    "cards",
		supportsCursor: true,
	},
	"gift_cards": {
		path:           "/gift-cards",
		responseKey:    "gift_cards",
		supportsLimit:  true,
		supportsCursor: true,
	},
	"payouts": {
		path:              "/payouts",
		responseKey:       "payouts",
		supportsLimit:     true,
		supportsCursor:    true,
		supportsTimeRange: true,
	},
	"disputes": {
		path:           "/disputes",
		responseKey:    "disputes",
		supportsCursor: true,
	},
	"bank_accounts": {
		path:           "/bank-accounts",
		responseKey:    "bank_accounts",
		supportsLimit:  true,
		supportsCursor: true,
	},
	"merchants": {
		path:           "/merchants",
		responseKey:    "merchant",
		supportsCursor: true,
	},
	"bookings/custom_attribute_definitions": {
		path:           "/bookings/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"bookings": {
		path:           "/bookings",
		responseKey:    "bookings",
		supportsCursor: true,
		supportsLimit:  true,
	},

	"bookings/location_booking_profiles": {
		path:           "/bookings/location-booking-profiles",
		responseKey:    "location_booking_profiles",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"bookings/team_member_booking_profiles": {
		path:           "/bookings/team-member-booking-profiles",
		responseKey:    "team_member_booking_profiles",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"channels": {
		path:           "/channels",
		responseKey:    "channels",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"online_checkout/payment_links": {
		path:           "/online-checkout/payment-links",
		responseKey:    "payment_links",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"customers/custom_attribute_definitions": {
		path:           "/customers/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"customers/groups": {
		path:           "/customers/groups",
		responseKey:    "groups",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"customers/segments": {
		path:           "/customers/segments",
		responseKey:    "segments",
		supportsCursor: true,
		supportsLimit:  true,
	},

	"devices": {
		path:           "/devices",
		responseKey:    "devices",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"devices/codes": {
		path:           "/devices/codes",
		responseKey:    "device_codes",
		supportsCursor: true,
		supportsLimit:  false,
	},

	"events/types": {
		path:           "/events/types",
		responseKey:    "event_types",
		supportsCursor: false,
		supportsLimit:  false,
	},

	"gift_cards/activities": {
		path:              "/gift-cards/activities",
		responseKey:       "gift_card_activities",
		supportsCursor:    true,
		supportsLimit:     true,
		supportsTimeRange: true,
	},

	"labor/break_types": {
		path:           "/labor/break-types",
		responseKey:    "break_types",
		supportsCursor: true,
		supportsLimit:  true,
	},

	"labor/team_member_wages": {
		path:           "/labor/team-member-wages",
		responseKey:    "team_member_wages",
		supportsCursor: true,
		supportsLimit:  true,
	},

	"labor/workweek_configs": {
		path:           "/labor/workweek-configs",
		responseKey:    "workweek_configs",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"locations/custom_attribute_definitions": {
		path:           "/locations/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"loyalty/programs": {
		path:        "/loyalty/programs",
		responseKey: "programs",
	},
	"merchants/custom_attribute_definitions": {
		path:           "/merchants/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"orders/custom_attribute_definitions": {
		path:           "/orders/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
	},
	"sites": {
		path:        "/sites",
		responseKey: "sites",
	},
	"team_members/jobs": {
		path:           "/team-members/jobs",
		responseKey:    "jobs",
		supportsCursor: true,
	},
	"webhooks/event_types": {
		path:        "/webhooks/event-types",
		responseKey: "event_types",
	},
	"webhooks/subscriptions": {
		path:           "/webhooks/subscriptions",
		responseKey:    "subscriptions",
		supportsCursor: true,
		supportsLimit:  true,
	},
}
