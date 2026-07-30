package square

import "net/http"

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

// writeConfig describes how to create and/or update an object's records on the
// Square API.
type writeConfig struct {
	// createPath is the URL path (after the base URL and API version) that
	// creates a record, e.g. "/customers". Empty means create is unsupported.
	createPath string
	// updatePath is the URL path that updates a record. The record id is
	// appended to it unless updateIDInBody is set. Empty means update is
	// unsupported.
	updatePath string
	// createKey and updateKey name the JSON envelope wrapping the record in the
	// request body, e.g. {"gift_card": {...}}. Empty sends the record fields at
	// the top level of the body.
	createKey string
	updateKey string
	// responseKey is the JSON key in the response that holds the written record.
	responseKey string
	// idField is the record field holding its identifier. Empty defaults to "id".
	idField string
	// createNeedsIdempotency and updateNeedsIdempotency report whether the
	// endpoint requires a top-level idempotency_key. When the caller omits it,
	// the connector generates a UUID.
	createNeedsIdempotency bool
	updateNeedsIdempotency bool
	// topLevelFields are record fields the endpoint expects as siblings of the
	// envelope rather than inside it, e.g. source_id when creating a card. They
	// are hoisted out of the record before wrapping.
	topLevelFields []string
	// updateMethod overrides the HTTP method for updates. Empty defaults to PUT.
	updateMethod string
	// updateIDInBody injects the record id into the wrapped record instead of
	// appending it to the URL path. Used by the catalog upsert endpoint.
	updateIDInBody bool
}

// writeObjects is the set of objects the Square connector can create and/or
// update. Objects absent from this map are read-only on the Square API.
var writeObjects = map[string]writeConfig{ //nolint:gochecknoglobals
	// https://developer.squareup.com/reference/square/customers-api/create-customer
	"customers": {
		createPath:  "/customers",
		updatePath:  "/customers",
		responseKey: "customer",
	},
	// https://developer.squareup.com/reference/square/locations-api/create-location
	"locations": {
		createPath:  "/locations",
		updatePath:  "/locations",
		createKey:   "location",
		updateKey:   "location",
		responseKey: "location",
	},
	// https://developer.squareup.com/reference/square/payments-api/create-payment
	// CreatePayment takes a flat body; UpdatePayment wraps the record in
	// "payment". Both require an idempotency key.
	"payments": {
		createPath:             "/payments",
		updatePath:             "/payments",
		updateKey:              "payment",
		responseKey:            "payment",
		createNeedsIdempotency: true,
		updateNeedsIdempotency: true,
	},
	// https://developer.squareup.com/reference/square/refunds-api/refund-payment
	"refunds": {
		createPath:             "/refunds",
		responseKey:            "refund",
		createNeedsIdempotency: true,
	},
	// https://developer.squareup.com/reference/square/catalog-api/upsert-catalog-object
	// The upsert endpoint both creates and updates: updates POST to the same
	// path with the record id (set from RecordId) and current version inside
	// the object.
	"catalog": {
		createPath:             "/catalog/object",
		updatePath:             "/catalog/object",
		createKey:              "object",
		updateKey:              "object",
		responseKey:            "catalog_object",
		createNeedsIdempotency: true,
		updateNeedsIdempotency: true,
		updateMethod:           http.MethodPost,
		updateIDInBody:         true,
	},
	// https://developer.squareup.com/reference/square/cards-api/create-card
	// source_id and verification_token sit beside the "card" envelope.
	"cards": {
		createPath:             "/cards",
		createKey:              "card",
		responseKey:            "card",
		createNeedsIdempotency: true,
		topLevelFields:         []string{"source_id", "verification_token"},
	},
	// https://developer.squareup.com/reference/square/gift-cards-api/create-gift-card
	// location_id sits beside the "gift_card" envelope.
	"gift_cards": {
		createPath:             "/gift-cards",
		createKey:              "gift_card",
		responseKey:            "gift_card",
		createNeedsIdempotency: true,
		topLevelFields:         []string{"location_id"},
	},
	// https://developer.squareup.com/reference/square/gift-card-activities-api/create-gift-card-activity
	"gift_cards/activities": {
		createPath:             "/gift-cards/activities",
		createKey:              "gift_card_activity",
		responseKey:            "gift_card_activity",
		createNeedsIdempotency: true,
	},
	// https://developer.squareup.com/reference/square/devices-api/create-device-code
	"devices/codes": {
		createPath:             "/devices/codes",
		createKey:              "device_code",
		responseKey:            "device_code",
		createNeedsIdempotency: true,
	},
	// https://developer.squareup.com/reference/square/bookings-api/create-booking
	"bookings": {
		createPath:  "/bookings",
		updatePath:  "/bookings",
		createKey:   "booking",
		updateKey:   "booking",
		responseKey: "booking",
	},
	// https://developer.squareup.com/reference/square/booking-custom-attributes-api
	"bookings/custom_attribute_definitions": {
		createPath:  "/bookings/custom-attribute-definitions",
		updatePath:  "/bookings/custom-attribute-definitions",
		createKey:   "custom_attribute_definition",
		updateKey:   "custom_attribute_definition",
		responseKey: "custom_attribute_definition",
		idField:     "key",
	},
	// https://developer.squareup.com/reference/square/customer-custom-attributes-api
	"customers/custom_attribute_definitions": {
		createPath:  "/customers/custom-attribute-definitions",
		updatePath:  "/customers/custom-attribute-definitions",
		createKey:   "custom_attribute_definition",
		updateKey:   "custom_attribute_definition",
		responseKey: "custom_attribute_definition",
		idField:     "key",
	},
	// https://developer.squareup.com/reference/square/location-custom-attributes-api
	"locations/custom_attribute_definitions": {
		createPath:  "/locations/custom-attribute-definitions",
		updatePath:  "/locations/custom-attribute-definitions",
		createKey:   "custom_attribute_definition",
		updateKey:   "custom_attribute_definition",
		responseKey: "custom_attribute_definition",
		idField:     "key",
	},
	// https://developer.squareup.com/reference/square/merchant-custom-attributes-api
	"merchants/custom_attribute_definitions": {
		createPath:  "/merchants/custom-attribute-definitions",
		updatePath:  "/merchants/custom-attribute-definitions",
		createKey:   "custom_attribute_definition",
		updateKey:   "custom_attribute_definition",
		responseKey: "custom_attribute_definition",
		idField:     "key",
	},
	// https://developer.squareup.com/reference/square/order-custom-attributes-api
	"orders/custom_attribute_definitions": {
		createPath:  "/orders/custom-attribute-definitions",
		updatePath:  "/orders/custom-attribute-definitions",
		createKey:   "custom_attribute_definition",
		updateKey:   "custom_attribute_definition",
		responseKey: "custom_attribute_definition",
		idField:     "key",
	},
	// https://developer.squareup.com/reference/square/customer-groups-api/create-customer-group
	"customers/groups": {
		createPath:  "/customers/groups",
		updatePath:  "/customers/groups",
		createKey:   "group",
		updateKey:   "group",
		responseKey: "group",
	},
	// https://developer.squareup.com/reference/square/labor-api/create-break-type
	"labor/break_types": {
		createPath:  "/labor/break-types",
		updatePath:  "/labor/break-types",
		createKey:   "break_type",
		updateKey:   "break_type",
		responseKey: "break_type",
	},
	// https://developer.squareup.com/reference/square/labor-api/update-workweek-config
	// Workweek configs cannot be created, only updated.
	"labor/workweek_configs": {
		updatePath:  "/labor/workweek-configs",
		updateKey:   "workweek_config",
		responseKey: "workweek_config",
	},
	// https://developer.squareup.com/reference/square/checkout-api/create-payment-link
	// CreatePaymentLink takes a flat body; UpdatePaymentLink wraps the record
	// in "payment_link".
	"online_checkout/payment_links": {
		createPath:  "/online-checkout/payment-links",
		updatePath:  "/online-checkout/payment-links",
		updateKey:   "payment_link",
		responseKey: "payment_link",
	},
	// https://developer.squareup.com/reference/square/team-api/create-job
	"team_members/jobs": {
		createPath:  "/team-members/jobs",
		updatePath:  "/team-members/jobs",
		createKey:   "job",
		updateKey:   "job",
		responseKey: "job",
	},
	// https://developer.squareup.com/reference/square/webhook-subscriptions-api/create-webhook-subscription
	"webhooks/subscriptions": {
		createPath:  "/webhooks/subscriptions",
		updatePath:  "/webhooks/subscriptions",
		createKey:   "subscription",
		updateKey:   "subscription",
		responseKey: "subscription",
	},
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
