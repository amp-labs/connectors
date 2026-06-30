package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
	"github.com/amp-labs/connectors/providers/stripe/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

const objectNameCheckoutSessions = "checkout/sessions"

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

func main() { // nolint:funlen
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
			return endpoint
		})),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(http.MethodGet,
		api3.OrPathMatcher{
			// Usual Path matching where IDs are excluded and concrete endpoints are ignored too.
			api3.AndPathMatcher{
				api3.IDPathIgnorer{},
				api3.NewDenyPathStrategy(ignoreEndpoints),
			},
			// Properties from Line Items are added into CheckoutSession object.
			api3.CustomPathMatcher(func(path string) bool {
				return path == "/v1/checkout/sessions/{session}/line_items"
			}),
		},
		nil, displayNameOverride,
		arrayLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	checkoutSessionLineItems := make([]metadatadef.Field, 0)

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
			if objectName == "checkout/sessions/{session}/line_items" {
				checkoutSessionLineItems = append(checkoutSessionLineItems, field)
			} else {
				// Usual behaviour.
				schemas.Add("", objectName, object.DisplayName, urlPath, object.ResponseKey,
					utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
			}
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	// Once all objects are complete enhance Checkout Sessions with LineItems properties.
	addLineItems(schemas, checkoutSessionLineItems)

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func addLineItems(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any], items []metadatadef.Field) {
	for _, field := range items {
		// Enhance Checkout Session object with expandable fields.
		fieldName := fmt.Sprintf("$['line_items']['data'][*]['%v']", field.Name)

		fieldDisplayName := strings.ReplaceAll(field.Name, "_", " ")
		fieldDisplayName = api3.CapitalizeFirstLetterEveryWord(fieldDisplayName)
		fieldDisplayName = fmt.Sprintf("Line Item's %v", fieldDisplayName)

		fieldV2 := staticschema.FieldMetadataMapV2{
			fieldName: staticschema.FieldMetadata{
				DisplayName:  fieldDisplayName,
				ValueType:    utilsopenapi.GetFieldValueType(field),
				ProviderType: field.Type,
				Values:       utilsopenapi.GetFieldValueOptions(field),
			},
		}
		schemas.Add("", objectNameCheckoutSessions, "", "", "",
			fieldV2, nil, false)
	}

	schemas.Add("", objectNameCheckoutSessions, "", "", "",
		staticschema.FieldMetadataMapV2{
			"$['line_items']['has_more']": staticschema.FieldMetadata{
				DisplayName:  "LineItem's Has more",
				ValueType:    common.ValueTypeBoolean,
				ProviderType: "bool",
			},
			"$['line_items']['url']": staticschema.FieldMetadata{
				DisplayName:  "LineItem's URL next page",
				ValueType:    common.ValueTypeString,
				ProviderType: "string",
			},
		}, nil, false)
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
