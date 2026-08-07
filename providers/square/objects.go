package square

const apiVersion = "v2"

// objectConfig is the single source of truth for a Square object: how to list
// its records, and (when supported) how to create/update them. Write fields
// left at their zero value mean the object is read-only.
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
	// supportsWrite reports whether records can be written: created via
	// POST path and updated via PUT path/{id}.
	supportsWrite bool
	// writeKey is the JSON envelope wrapping the record in write request and
	// response bodies, e.g. {"gift_card": {...}}. Empty means flat bodies.
	writeKey string
	// writeResponseKey overrides writeKey as the response envelope, for objects
	// whose request body is flat but whose response is wrapped (customers,
	// refunds) or whose response key differs (catalog).
	writeResponseKey string
	// idField is the record field holding its identifier. Empty defaults to "id".
	idField string
	// needsIdempotency reports whether writes require a top-level
	// idempotency_key. When the caller omits it, the connector generates a UUID.
	needsIdempotency bool
	// topLevelFields are record fields the endpoint expects as siblings of the
	// envelope rather than inside it, e.g. source_id when creating a card.
	topLevelFields []string
	// upsertPath, when set, is where writes go instead of path, and marks the
	// endpoint as an upsert: updates POST there with the record id injected
	// into the body instead of PUT path/{id} (the catalog upsert endpoint).
	upsertPath string
}

var objects = map[string]objectConfig{ //nolint:gochecknoglobals
	// https://developer.squareup.com/reference/square/customers-api/create-customer
	// Create and update take flat bodies; the response wraps the record in "customer".
	"customers": {
		path:             "/customers",
		responseKey:      "customers",
		supportsLimit:    true,
		supportsCursor:   true,
		supportsWrite:    true,
		writeResponseKey: "customer",
	},
	// https://developer.squareup.com/reference/square/locations-api/create-location
	"locations": {
		path:          "/locations",
		responseKey:   "locations",
		supportsWrite: true,
		writeKey:      "location",
	},
	// https://developer.squareup.com/reference/square/payments-api/create-payment
	// CreatePayment takes a flat body; UpdatePayment wraps the record in "payment".
	"payments": {
		path:              "/payments",
		responseKey:       "payments",
		supportsLimit:     true,
		supportsCursor:    true,
		supportsTimeRange: true,
		supportsWrite:     true,
		writeKey:          "payment",
		needsIdempotency:  true,
	},
	// https://developer.squareup.com/reference/square/refunds-api/refund-payment
	"refunds": {
		path:              "/refunds",
		responseKey:       "refunds",
		supportsLimit:     true,
		supportsCursor:    true,
		supportsTimeRange: true,
		supportsWrite:     true,
		writeResponseKey:  "refund",
		needsIdempotency:  true,
	},
	// https://developer.squareup.com/reference/square/catalog-api/upsert-catalog-object
	// The upsert endpoint both creates and updates: updates POST to the same
	// path with the record id (set from RecordId) and current version inside
	// the object.
	"catalog": {
		path:             "/catalog/list",
		responseKey:      "objects",
		supportsCursor:   true,
		supportsWrite:    true,
		upsertPath:       "/catalog/object",
		writeKey:         "object",
		writeResponseKey: "catalog_object",
		needsIdempotency: true,
	},
	// https://developer.squareup.com/reference/square/cards-api/create-card
	// source_id and verification_token sit beside the "card" envelope.
	"cards": {
		path:             "/cards",
		responseKey:      "cards",
		supportsCursor:   true,
		supportsWrite:    true,
		writeKey:         "card",
		needsIdempotency: true,
		topLevelFields:   []string{"source_id", "verification_token"},
	},
	// https://developer.squareup.com/reference/square/gift-cards-api/create-gift-card
	// location_id sits beside the "gift_card" envelope.
	"gift_cards": {
		path:             "/gift-cards",
		responseKey:      "gift_cards",
		supportsLimit:    true,
		supportsCursor:   true,
		supportsWrite:    true,
		writeKey:         "gift_card",
		needsIdempotency: true,
		topLevelFields:   []string{"location_id"},
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
	// https://developer.squareup.com/reference/square/booking-custom-attributes-api
	"bookings/custom_attribute_definitions": {
		path:           "/bookings/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "custom_attribute_definition",
		idField:        "key",
	},
	// https://developer.squareup.com/reference/square/bookings-api/create-booking
	"bookings": {
		path:           "/bookings",
		responseKey:    "bookings",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "booking",
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
	// https://developer.squareup.com/reference/square/checkout-api/create-payment-link
	// CreatePaymentLink takes a flat body; UpdatePaymentLink wraps the record
	// in "payment_link".
	"online_checkout/payment_links": {
		path:           "/online-checkout/payment-links",
		responseKey:    "payment_links",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "payment_link",
	},
	// https://developer.squareup.com/reference/square/customer-custom-attributes-api
	"customers/custom_attribute_definitions": {
		path:           "/customers/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "custom_attribute_definition",
		idField:        "key",
	},
	// https://developer.squareup.com/reference/square/customer-groups-api/create-customer-group
	"customers/groups": {
		path:           "/customers/groups",
		responseKey:    "groups",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "group",
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
	// https://developer.squareup.com/reference/square/devices-api/create-device-code
	"devices/codes": {
		path:             "/devices/codes",
		responseKey:      "device_codes",
		supportsCursor:   true,
		supportsLimit:    false,
		supportsWrite:    true,
		writeKey:         "device_code",
		needsIdempotency: true,
	},

	"events/types": {
		path:           "/events/types",
		responseKey:    "event_types",
		supportsCursor: false,
		supportsLimit:  false,
	},

	// https://developer.squareup.com/reference/square/gift-card-activities-api/create-gift-card-activity
	"gift_cards/activities": {
		path:              "/gift-cards/activities",
		responseKey:       "gift_card_activities",
		supportsCursor:    true,
		supportsLimit:     true,
		supportsTimeRange: true,
		supportsWrite:     true,
		writeKey:          "gift_card_activity",
		needsIdempotency:  true,
	},

	// https://developer.squareup.com/reference/square/labor-api/create-break-type
	"labor/break_types": {
		path:           "/labor/break-types",
		responseKey:    "break_types",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "break_type",
	},

	"labor/team_member_wages": {
		path:           "/labor/team-member-wages",
		responseKey:    "team_member_wages",
		supportsCursor: true,
		supportsLimit:  true,
	},

	// https://developer.squareup.com/reference/square/labor-api/update-workweek-config
	"labor/workweek_configs": {
		path:           "/labor/workweek-configs",
		responseKey:    "workweek_configs",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "workweek_config",
	},
	// https://developer.squareup.com/reference/square/location-custom-attributes-api
	"locations/custom_attribute_definitions": {
		path:           "/locations/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "custom_attribute_definition",
		idField:        "key",
	},
	"loyalty/programs": {
		path:        "/loyalty/programs",
		responseKey: "programs",
	},
	// https://developer.squareup.com/reference/square/merchant-custom-attributes-api
	"merchants/custom_attribute_definitions": {
		path:           "/merchants/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "custom_attribute_definition",
		idField:        "key",
	},
	// https://developer.squareup.com/reference/square/order-custom-attributes-api
	"orders/custom_attribute_definitions": {
		path:           "/orders/custom-attribute-definitions",
		responseKey:    "custom_attribute_definitions",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "custom_attribute_definition",
		idField:        "key",
	},
	"sites": {
		path:        "/sites",
		responseKey: "sites",
	},
	// https://developer.squareup.com/reference/square/team-api/create-job
	"team_members/jobs": {
		path:           "/team-members/jobs",
		responseKey:    "jobs",
		supportsCursor: true,
		supportsWrite:  true,
		writeKey:       "job",
	},
	"webhooks/event_types": {
		path:        "/webhooks/event-types",
		responseKey: "event_types",
	},
	// https://developer.squareup.com/reference/square/webhook-subscriptions-api/create-webhook-subscription
	"webhooks/subscriptions": {
		path:           "/webhooks/subscriptions",
		responseKey:    "subscriptions",
		supportsCursor: true,
		supportsLimit:  true,
		supportsWrite:  true,
		writeKey:       "subscription",
	},
}
