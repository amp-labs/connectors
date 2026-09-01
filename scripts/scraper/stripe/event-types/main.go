// nolint:forbidigo
package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var objectPrefixToAmpersandObjectName = datautils.Map[string, string]{ // nolint:gochecknoglobals
	"account":                            "accounts",
	"application_fee":                    "application_fees",
	"billing.alert":                      "billing/alerts",
	"billing.credit_balance_transaction": "billing/credit_balance_transactions",
	"billing.credit_grant":               "billing/credit_grants",
	"billing.meter":                      "billing/meters",
	"billing_portal.configuration":       "billing_portal/configurations",
	"charge":                             "charges",
	"charge.dispute":                     "disputes",
	"charge.refund":                      "refunds",
	"checkout.session":                   "checkout/sessions",
	"climate.order":                      "climate/orders",
	"climate.product":                    "climate/products",
	"coupon":                             "coupons",
	"credit_note":                        "credit_notes",
	"customer":                           "customers",
	"customer.subscription":              "subscriptions",
	"customer.tax_id":                    "tax_ids",
	"file":                               "files",
	"financial_connections.account":      "financial_connections/accounts",
	"identity.verification_session":      "identity/verification_sessions",
	"invoice":                            "invoices",
	"invoice_payment":                    "invoice_payments",
	"invoiceitem":                        "invoiceitems",
	"issuing_authorization":              "issuing/authorizations",
	"issuing_card":                       "issuing/cards",
	"issuing_cardholder":                 "issuing/cardholders",
	"issuing_dispute":                    "issuing/disputes",
	"issuing_personalization_design":     "issuing/personalization_designs",
	"issuing_transaction":                "issuing/transactions",
	"payment_intent":                     "payment_intents",
	"payment_link":                       "payment_links",
	"payment_method":                     "payment_methods",
	"payout":                             "payouts",
	"plan":                               "plans",
	"price":                              "prices",
	"product":                            "products",
	"promotion_code":                     "promotion_codes",
	"quote":                              "quotes",
	"radar.early_fraud_warning":          "radar/early_fraud_warnings",
	"refund":                             "refunds",
	"reporting.report_run":               "reporting/report_runs",
	"reporting.report_type":              "reporting/report_types",
	"review":                             "reviews",
	"setup_intent":                       "setup_intents",
	"sigma.scheduled_query_run":          "sigma/scheduled_query_runs",
	"subscription_schedule":              "subscription_schedules",
	"tax_rate":                           "tax_rates",
	"terminal.reader":                    "terminal/readers",
	"test_helpers.test_clock":            "test_helpers/test_clocks",
	"topup":                              "topups",
	"transfer":                           "transfers",
}

func main() {
	var events []string

	goutils.MustBeNil(scrapper.LoadFile("scripts/scraper/stripe/event-types/events.json", &events))

	registry := make(datautils.IndexedLists[string, string])

	for _, event := range events {
		i := strings.LastIndex(event, ".")
		prefix := event[:i]
		operation := event[i+1:]

		registry.Add(prefix, operation)
	}

	prefixes := registry.GetBuckets()
	sort.Strings(prefixes)

	fmt.Println("===== All Events =====")

	for _, prefix := range prefixes {
		objectEvents := registry[prefix]

		fmt.Println(prefix)

		for _, event := range objectEvents {
			fmt.Printf("\t%v\n", event)
		}
	}

	fmt.Println("===== Excluded Event prefixes =====")

	excludedPrefixes := datautils.NewSetFromList(prefixes).Subtract(objectPrefixToAmpersandObjectName.KeySet())
	sort.Strings(excludedPrefixes)

	for _, prefix := range excludedPrefixes {
		fmt.Println(prefix)
	}

	type Object map[string]string

	output := make(map[string]Object)

	for prefix, objectName := range objectPrefixToAmpersandObjectName {
		objectEvents := registry[prefix]

		object := make(Object)
		for index, suffix := range objectEvents {
			// Must be filled in manually.
			// Manually map Create/Update/Delete Ampersand events.
			object[strconv.Itoa(index)] = prefix + "." + suffix
		}

		output[objectName] = object
	}

	goutils.MustBeNil(fileconv.Flusher{}.ToFile("scripts/scraper/stripe/event-types/mapping.json", output))
}
