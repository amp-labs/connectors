package main

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/stripe/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// not a list resource, rather singular object
		"/v1/account",           // account details
		"/v1/balance",           // current balance associated with auth
		"/v1/invoices/upcoming", // pending charges, invoices, subscriptions for a customer
		"/v1/tax/settings",      // tax settings for a merchant
		"/v1/balance_settings",  // holds list of enums aka balance settings
		// Search endpoints are similar and covered by related endpoints.
		"*/search",
		// Required query params when making API call.
		"/v1/invoices/upcoming/lines",
		// Endpoints missing from official documentation. Sending GET request gives an error.
		"/v1/exchange_rates",
		"/v1/linked_accounts",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"v1/financial_connections/accounts": "Financial Connection Accounts",
		"v1/terminal/configurations":        "Terminal Configurations",
		"history":                           "Balance History Transactions",
		"invoices":                          "Invoices",
		"locations":                         "Terminal Locations",
		"payment_intents":                   "Payment Intents",
		"payment_methods":                   "Payment Methods",
		"refunds":                           "API Method Refunds",
		"report_runs":                       "Financial Report Runs",
		"report_types":                      "Financial Report Types",
		"v1/checkout/sessions":              "Payment Checkout Sessions",
		"setup_intents":                     "Payment Setup Intents",
		"subscriptions":                     "Subscriptions",
		"suppliers":                         "Climate Suppliers",
		"tax_ids":                           "Tax Identifiers",
		"topups":                            "Top-ups",
		"value_lists":                       "Radar Value Lists",
		"verification_reports":              "Verification Reports",
		"verification_sessions":             "Verification Sessions",
		"webhook_endpoints":                 "Webhook Endpoints",
	}
)

func main() { // nolint:funlen
	explorer, err := files.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			removeListSuffix,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			processResourceTitles,
			api3.Pluralize,
		),
		api3.WithParameterFilterGetMethod(
			OnlyOptionalQueryParametersOrTreasuryApi,
		),
		api3.WithArrayItemAutoSelection(),
		api3.WithDuplicatesResolver(api3.SingleItemDuplicatesResolver(func(endpoint string) string {
			return endpoint
		})),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(http.MethodGet,
		api3.AndPathMatcher{
			api3.IDPathIgnorer{},
			api3.NewDenyPathStrategy(ignoreEndpoints),
		},
		nil, displayNameOverride,
		arrayLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		urlPath, _ := strings.CutPrefix(object.URLPath, "/v1/")
		objectName := urlPath

		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", objectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			fieldMetadata := staticschema.FieldMetadataMapV2{
				field.Name: staticschema.FieldMetadata{
					DisplayName:  formatFieldDisplayName(field.Name),
					ValueType:    utilsopenapi.GetFieldValueType(field),
					ProviderType: field.Type,
					Values:       utilsopenapi.GetFieldValueOptions(field),
				},
			}
			schemas.Add("", objectName, object.DisplayName, urlPath, object.ResponseKey,
				fieldMetadata, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	goutils.MustBeNil(files.OutputStripe.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputStripe.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func formatFieldDisplayName(fieldName string) string {
	displayName := fieldName
	displayName = naming.SeparateUnderscoreWords(displayName)
	displayName = naming.CapitalizeFirstLetterEveryWord(displayName)

	return displayName
}

func arrayLocator(objectName, fieldName string) bool {
	slog.Warn("unexpected call to locator, provider API was expected to have no ambiguous array fields",
		"object", objectName)

	return false
}

func removeListSuffix(displayName string) string {
	result, _ := strings.CutSuffix(displayName, "List")

	return result
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

func OnlyOptionalQueryParametersOrTreasuryApi(urlPath string, operation *openapi3.Operation) bool {
	for _, parameter := range operation.Parameters {
		// Required query params are not supported.
		if parameter.Value.In == "query" && parameter.Value.Required {
			// Treasury endpoints are an exception.
			if strings.Contains(urlPath, "/v1/treasury/") &&
				strings.Contains(parameter.Value.Name, "financial_account") {
				return true
			}

			return false
		}
	}

	return true
}
