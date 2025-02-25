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
		"financial_connections/accounts": "Financial Connection Accounts",
		"terminal/configurations":        "Terminal Configurations",
		"history":                        "Balance History Transactions",
		"invoices":                       "Invoices",
		"locations":                      "Terminal Locations",
		"payment_intents":                "Payment Intents",
		"payment_methods":                "Payment Methods",
		"refunds":                        "API Method Refunds",
		"report_runs":                    "Financial Report Runs",
		"report_types":                   "Financial Report Types",
		"checkout/sessions":              "Payment Checkout Sessions",
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
		api3.WithDuplicatesResolver(api3.SingleItemDuplicatesResolver(func(endpoint string) string {
			objectName, _ := strings.CutPrefix(endpoint, "/v1/")

			return objectName
		})),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride,
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
				}, nil, object.Custom)
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
