package highlevelstandard

import (
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "2021-07-28"
	defaultPageSize = 100
)

var objectsWithLocationIdInParam = datautils.NewSet( //nolint:gochecknoglobals
	"businesses",
	"calendars",
	"calendars/groups",
	"campaigns",
	"contacts",
	"conversations/search",
	"emails/schedule",
	"forms/submissions",
	"forms",
	"links",
	"blogs/authors",
	"blogs/categories",
	"funnels/lookup/redirect/list",
	"funnels/funnel/list",
	"opportunities/pipelines",
	"products",
	"proposals/document",
	"proposals/templates",
	"surveys",
	"users",
	"workflows",
	"objects",
)

var paginationObjects = datautils.NewSet( //nolint:gochecknoglobals
	"emails/schedule",
	"invoices",
	"invoices/template",
	"invoices/schedule",
	"invoices/estimate/list",
	"invoices/estimate/template",
	"blogs/authors",
	"blogs/categories",
	"funnels/lookup/redirect/list",
	"funnels/funnel/list",
	"payments/orders",
	"payments/transactions",
	"payments/subscriptions",
	"payments/coupon/list",
	"products",
	"products/inventory",
	"products/collections",
	"products/reviews",
	"store/shipping-zone",
	"locations/search",
	"custom-menus",
	"surveys",
	"proposals/document",
	"proposals/templates",
	"marketplace/billing/charges",
	"forms",
)

var objectWithAltTypeAndIdQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"invoices",
	"invoices/template",
	"invoices/schedule",
	"invoices/estimate/list",
	"invoices/estimate/template",
	"payments/orders",
	"payments/transactions",
	"payments/subscriptions",
	"payments/coupon/list",
	"products/inventory",
	"products/collections",
	"products/reviews",
	"store/shipping-zone",
	"store/shipping-carrier",
	"store/store-setting",
)

var objectWithSkipQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"custom-menus",
	"locations/search",
	"proposals/document",
	"proposals/templates",
	"surveys",
	"invoices",
	"marketplace/billing/charges",
)

// Ref for nodePath https://highlevel.stoplight.io/docs/integrations/a8db8afcbe0a3-get-businesses-by-location.
var objectsNodePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"businesses":                   "businesses",
	"calendars":                    "calendars",
	"calendars/groups":             "groups",
	"campaigns":                    "campaigns",
	"contacts":                     "contacts",
	"conversations/search":         "conversations",
	"emails/schedule":              "schedules",
	"forms/submissions":            "submissions",
	"forms":                        "forms",
	"invoices":                     "invoices",
	"invoices/template":            "data",
	"invoices/schedule":            "schedules",
	"invoices/estimate/list":       "estimates",
	"invoices/estimate/template":   "data",
	"links":                        "links",
	"blogs/authors":                "authors",
	"blogs/categories":             "categories",
	"funnels/lookup/redirect/list": "data",
	"funnels/funnel/list":          "funnels",
	"opportunities/pipelines":      "pipelines",
	"payments/orders":              "data",
	"payments/transactions":        "data",
	"payments/subscriptions":       "data",
	"payments/coupon/list":         "data",
	"products":                     "products",
	"products/inventory":           "inventory",
	"products/collections":         "data",
	"products/reviews":             "data",
	"proposals/document":           "documents",
	"proposals/templates":          "data",
	"store/shipping-zone":          "data",
	"store/shipping-carrier":       "data",
	"store/store-setting":          "data",
	"surveys":                      "surveys",
	"users":                        "users ",
	"workflows":                    "workflows",
	"locations/search":             "locations",
	"custom-menus":                 "customMenus",
	"marketplace/billing/charges":  "charges",
	"objects":                      "objects",
}, func(objectName string) string {
	return "id"
},
)

// makeNextRecord creates a function that determines the next page token based on the current offset.
func makeNextRecord(offset int, objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if !paginationObjects.Has(objName) {
			return "", nil
		}

		nextStart := offset + defaultPageSize

		return strconv.Itoa(nextStart), nil
	}
}

// Objects paginated with the startAfterId/startAfter cursor returned in the response "meta" object.
// Ref https://highlevel.stoplight.io/docs/integrations/ab55933a57f6f-get-contacts.
var cursorPaginationObjects = datautils.NewSet( //nolint:gochecknoglobals
	"contacts",
)

const (
	cursorParamStartAfterId = "startAfterId"
	cursorParamStartAfter   = "startAfter"
)

// makeNextCursor builds the next page token from the response "meta" object.
// The token is an encoded query string holding the startAfterId and startAfter values.
// An empty token is returned when the response has no next page.
func makeNextCursor(node *ajson.Node) (string, error) {
	meta, err := jsonquery.New(node).ObjectOptional("meta")
	if err != nil || meta == nil {
		return "", err
	}

	nextPageURL, err := jsonquery.New(meta).StrWithDefault("nextPageUrl", "")
	if err != nil {
		return "", err
	}

	startAfterId, err := jsonquery.New(meta).StrWithDefault(cursorParamStartAfterId, "")
	if err != nil {
		return "", err
	}

	if nextPageURL == "" || startAfterId == "" {
		return "", nil
	}

	startAfter, err := jsonquery.New(meta).IntegerWithDefault(cursorParamStartAfter, 0)
	if err != nil {
		return "", err
	}

	cursor := url.Values{}
	cursor.Set(cursorParamStartAfterId, startAfterId)
	cursor.Set(cursorParamStartAfter, strconv.FormatInt(startAfter, 10))

	return cursor.Encode(), nil
}

var writeObjectsNodePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"custom-menus":                    "customMenu",
	"users":                           "",
	"businesses":                      "business",
	"calendars":                       "calendar",
	"calendars/groups":                "group",
	"calendars/events/appointments":   "",
	"calendars/events/block-slots":    "",
	"contacts":                        "contact",
	"objects":                         "object",
	"associations":                    "",
	"associations/relations":          "",
	"custom-fields":                   "field",
	"custom-fields/folder":            "",
	"conversations":                   "conversation",
	"conversations/messages":          "",
	"conversations/messages/inbound":  "",
	"conversations/messages/outbound": "",
	"conversations/messages/upload":   "",
	"emails/builder":                  "",
	"invoices":                        "",
	"invoices/template":               "",
	"invoices/schedule":               "",
	"invoices/text2pay":               "invoice",
	"invoices/estimate":               "",
	"invoices/estimate/template":      "",
	"links":                           "link",
	"locations":                       "",
	"blogs/posts":                     "data",
	"funnels/lookup/redirect":         "data",
	"opportunities":                   "opportunity",
	"payments/coupon":                 "",
	"products":                        "",
	"products/collections":            "data",
	"store/shipping-zone":             "data",
	"store/shipping-carrier":          "data",
	"store/store-setting":             "data",
}, func(objectName string) string {
	return "id"
})

var writeObjectsWithIdField = datautils.NewSet( //nolint:gochecknoglobals
	"custom-menus",
	"users",
	"businesses",
	"calendars",
	"calendars/groups",
	"calendars/events/appointments",
	"calendars/events/block-slots",
	"contacts",
	"objects",
	"associations",
	"associations/relations",
	"custom-fields",
	"custom-fields/folder",
	"conversations",
	"links",
	"locations",
	"funnels/lookup/redirect",
	"opportunities",
)

var writeObjectsWithUnderscoreIdField = datautils.NewSet( //nolint:gochecknoglobals
	"invoices",
	"invoices/template",
	"invoices/schedule",
	"invoices/text2pay",
	"invoices/estimate",
	"invoices/estimate/template",
	"blogs/posts",
	"payments/coupon",
	"products",
	"products/collections",
	"store/shipping-zone",
	"store/shipping-carrier",
	"store/store-setting",
)

var deleteObjectWithLocationIdQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"associations/relations",
	"custom-fields/folder",
	"funnels/lookup/redirect",
)

var objectWithAltTypeAndIdBodyParam = datautils.NewSet( //nolint:gochecknoglobals
	"invoices/estimate",
	"invoices/estimate/template",
)
