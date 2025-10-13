package chargebee

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

var objectResponseField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"subscriptions":               "subscription",
	"customers":                   "customer",
	"invoices":                    "invoice",
	"orders":                      "order",
	"plans":                       "plan",
	"events":                      "event",
	"comments":                    "comment",
	"webhook_endpoints":           "webhook_endpoint",
	"currencies":                  "currency",
	"payment_sources":             "payment_source",
	"virtual_bank_accounts":       "virtual_bank_account",
	"transactions":                "transaction",
	"hosted_pages":                "hosted_page",
	"features":                    "feature",
	"entitlements":                "entitlement",
	"item_families":               "item_family",
	"items":                       "item",
	"item_prices":                 "item_price",
	"attached_items":              "attached_item",
	"coupons":                     "coupon",
	"coupon_sets":                 "coupon_set",
	"coupon_codes":                "coupon_code",
	"quotes":                      "quote",
	"credit_notes":                "credit_note",
	"promotional_credits":         "promotional_credit",
	"unbilled_charges":            "unbilled_charge",
	"omnichannel_subscriptions":   "omnichannel_subscription",
	"omnichannel_one_time_orders": "omnichannel_one_time_order",
	"usages":                      "usage",
	"gifts":                       "gift",
	"business_entities/transfers": "business_entity_transfer",
}, func(objectName string) string {
	return objectName
})

var objectNameWithListSuffix = datautils.NewSet[string]( //nolint:gochecknoglobals
	"currencies",
)
