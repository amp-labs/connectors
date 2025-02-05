package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
	"github.com/amp-labs/connectors/providers/stripe/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// not a list resource, rather singular object
		"/v1/account",           // account details
		"/v1/balance",           // current balance associated with auth
		"/v1/invoices/upcoming", // pending charges, invoices, subscriptions for a customer
		"/v1/tax/settings",      // tax settings for a merchant
		// Search endpoints are similar and covered by related endpoints.
		"*/search",
		// Required query params when making API call.
		"/v1/invoices/upcoming/lines",
		// Endpoints missing from official documentation. Sending GET request gives an error.
		"/v1/exchange_rates",
		"/v1/linked_accounts",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"accounts_financial_connections": "Financial Connection Accounts",
		"configurations_terminal":        "Terminal Configurations",
		"history":                        "Balance History Transactions",
		"invoices":                       "Invoices",
		"locations":                      "Terminal Locations",
		"payment_intents":                "Payment Intents",
		"payment_methods":                "Payment Methods",
		"refunds":                        "API Method Refunds",
		"report_runs":                    "Financial Report Runs",
		"report_types":                   "Financial Report Types",
		"sessions_checkout":              "Payment Checkout Sessions",
		"setup_intents":                  "Payment Setup Intents",
		"subscriptions":                  "Subscriptions",
		"suppliers":                      "Climate Suppliers",
		"tax_ids":                        "Tax Identifiers",
		"topups":                         "Top-ups",
		"value_lists":                    "Radar Value Lists",
		"verification_reports":           "Verification Reports",
		"verification_sessions":          "Verification Sessions",
		"webhook_endpoints":              "Webhook Endpoints",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		// Accounts
		"/v1/accounts":                       "accounts",
		"/v1/financial_connections/accounts": "accounts_financial_connections",
		// Authorizations
		"/v1/issuing/authorizations":              "authorizations",
		"/v1/test_helpers/issuing/authorizations": "authorizations_test",
		// Configurations
		"/v1/billing_portal/configurations": "configurations",
		"/v1/terminal/configurations":       "configurations_terminal",
		// Disputes
		"/v1/disputes":         "disputes",
		"/v1/issuing/disputes": "disputes_issuing",
		// Lines
		"/v1/credit_notes/preview/lines": "lines_preview_credit_notes",
		"/v1/invoices/upcoming/lines":    "lines_upcoming_invoices",
		// Products
		"/v1/climate/products": "products_climate",
		"/v1/products":         "products",
		// Received debits
		"/v1/test_helpers/treasury/received_debits": "received_debits_test",
		"/v1/treasury/received_debits":              "received_debits",
		// Received credits
		"/v1/test_helpers/treasury/received_credits": "received_credits_test",
		"/v1/treasury/received_credits":              "received_credits",
		// Sessions
		"/v1/billing_portal/sessions":        "sessions_billing_portal",
		"/v1/checkout/sessions":              "sessions_checkout",
		"/v1/financial_connections/sessions": "sessions_financial_connections",
		// Tokens
		"/v1/issuing/tokens": "tokens_issuing",
		"/v1/tokens":         "tokens",
		// Transactions
		"/v1/financial_connections/transactions": "transactions_financial_connections",
		"/v1/issuing/transactions":               "transactions_issuing",
		"/v1/treasury/transactions":              "transactions_treasury",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			removeListSuffix,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			processResourceTitles,
			api3.Pluralize,
		),
		api3.WithParameterFilterGetMethod(
			api3.OnlyOptionalQueryParameters,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		arrayLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field.Name: field.Name,
				}, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func arrayLocator(objectName, fieldName string) bool {
	slog.Warn("unexpected call to locator, provider API was expected to have no ambiguous array fields",
		"object", objectName)

	return false
}

func removeListSuffix(displayName string) string {
	output, _ := strings.CutSuffix(displayName, "List")

	return output
}

func processResourceTitles(displayName string) string {
	// Object "test_clocks" has a resource title of BillingClocksResourceBillingClockList.
	// There are several objects like this. We programmatically would pretend that title is BillingClock.
	for _, middleSeparator := range []string{
		" Resource ",
		" Resources ",
	} {
		parts := strings.Split(displayName, middleSeparator)
		if len(parts) == 2 { // nolint:mnd
			return parts[1]
		}
	}

	return displayName
}
