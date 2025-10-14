package chargebee

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
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

var objectNameWithListSuffix = datautils.NewSet( //nolint:gochecknoglobals
	"currencies",
)

//
//nolint:gochecknoglobals
var supportIncrementalRead = datautils.NewStringSet(
	"coupons",
	"credit_notes",
	"customers",
	"hosted_pages",
	"invoices",
	"item_prices",
	"items",
	"orders",
	"payment_sources",
	"quotes",
	"subscriptions",
	"transactions",
	"usages",
	"virtual_bank_accounts",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"attached_items",
		"business_entities/transfers",
		"comments",
		"coupon_codes",
		"coupon_sets",
		"coupons",
		"credit_notes",
		"currencies",
		"customers",
		"entitlements",
		"events",
		"features",
		"gifts",
		"hosted_pages",
		"invoices",
		"item_families",
		"item_prices",
		"items",
		"omnichannel_one_time_orders",
		"omnichannel_subscriptions",
		"orders",
		"payment_sources",
		"plans",
		"promotional_credits",
		"quotes",
		"subscriptions",
		"transactions",
		"unbilled_charges",
		"usages",
		"virtual_bank_accounts",
		"webhook_endpoints",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
