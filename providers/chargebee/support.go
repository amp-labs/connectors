package chargebee

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Api docs: https://apidocs.chargebee.com/docs/api/
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
	"recorded_purchases":          "recorded_purchase",
	"payment_schedule_schemes":    "payment_schedule_scheme",
	"offer_fulfillments":          "offer_fulfillment",
	"portal_sessions":             "portal_session",
	"payment_sources/create_using_temp_token":            "payment_source",
	"payment_sources/create_using_permanent_token":       "payment_source",
	"payment_sources/create_using_token":                 "payment_source",
	"payment_sources/create_using_payment_intent":        "payment_source",
	"payment_sources/create_card":                        "payment_source",
	"payment_sources/create_bank_account":                "payment_source",
	"payment_sources/create_voucher_payment_source":      "payment_source",
	"payment_intents":                                    "payment_intent",
	"virtual_bank_accounts/create_using_permanent_token": "virtual_bank_account",
	"payment_vouchers":                                   "payment_voucher",
}, func(objectName string) string {
	return objectName
})

var objectNameWithListSuffix = datautils.NewSet( //nolint:gochecknoglobals
	"currencies",
)

// objectNameWrite maps clean resource names to their corresponding API endpoints.
// This is used only for resources that have action/verb-based endpoints but we want
// to provide a cleaner, resource-based interface to users.
// Example: "invoices" -> "invoices/create_for_charge_items_and_charges".
var objectNameWrite = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"invoices":            "invoices/create_for_charge_items_and_charges",
	"promotional_credits": "promotional_credits/add",
	"quotes":              "quotes/create_for_charge_items_and_charges",
	"estimates":           "estimates/create_subscription_for_items",
	"coupons":             "coupons/create_for_items",
	"transactions":        "transactions/create_authorization",
}, func(objectName string) string {
	return objectName
})

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

func supportedOperations() components.EndpointRegistryInput { //nolint:funlen
	// ChargeBee objects that support read operation
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

	// writeSupport contains a mix of:
	// 1. Mapped object names (e.g., "invoices" -> "invoices/create_for_charge_items_and_charges")
	// 2. Raw object names (e.g., "payment_schedule_schemes", "items")
	writeSupport := []string{
		"customers",
		"recorded_purchases",
		"invoices",
		"credit_notes",
		"promotional_credits",
		"unbilled_charges",
		"payment_schedule_schemes",
		"quotes",
		"estimates",
		"orders",
		"item_families",
		"items",
		"item_prices",
		"coupons",
		"coupon_sets",
		"offer_fulfillments",
		"offer_events",
		"features",
		"portal_sessions",
		"payment_intents",
		"virtual_bank_accounts",
		"transactions",
		"payment_vouchers",
		"currencies",
		"webhook_endpoints",
		"comments",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
