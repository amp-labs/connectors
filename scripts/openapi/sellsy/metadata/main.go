package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/scripts/openapi/sellsy/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Singular object.
		"*/metas",
		"/accounts/conformities",
		"/email/authenticate",
		"/quotas",
		"/scopes",
		"/scopes/tree",
		"/search",
		"/settings/accounting-charts",
		"/settings/email",
	}
	objectEndpoints = map[string]string{}
	searchEndpoints = []string{
		"*/search",
	}
	ignoreSearchEndpoints = []string{
		// Singular object.
		"/estimates/search",
	}
	searchObjectEndpoints = map[string]string{
		"/accounting-codes/search":                   "accounting-codes",
		"/activities/crm/search":                     "crm",
		"/activities/search":                         "activities",
		"/calendar-events/search":                    "calendar-events",
		"/comments/search":                           "comments",
		"/companies/search":                          "companies",
		"/contacts/search":                           "contacts",
		"/credit-notes/search":                       "credit-notes",
		"/custom-activities/search":                  "custom-activities",
		"/custom-fields/search":                      "custom-fields",
		"/deposit-invoices/search":                   "deposit-invoices",
		"/documents/models/search":                   "documents/models",
		"/individuals/search":                        "individuals",
		"/invoices/search":                           "invoices",
		"/items/barcodes/search":                     "barcodes",
		"/items/search":                              "items",
		"/mandates/search":                           "mandates",
		"/notifications/search":                      "notifications",
		"/ocr/pur-invoice/search":                    "pur-invoice",
		"/opportunities/pipelines/search":            "pipelines",
		"/opportunities/search":                      "opportunities",
		"/opportunities/sources/search":              "sources",
		"/opportunities/steps/search":                "steps",
		"/orders/search":                             "orders",
		"/payments/methods/search":                   "methods",
		"/payments/search":                           "payments",
		"/phone-calls/search":                        "phone-calls",
		"/proposals/models/search":                   "proposals/models",
		"/staffs/search":                             "staffs",
		"/subscriptions/payment-installments/search": "payment-installments",
		"/subscriptions/search":                      "subscriptions",
		"/tasks/search":                              "tasks",
		"/taxes/search":                              "taxes",
		"/webhooks/search":                           "webhooks",
	}
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	objects := Objects()
	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(files.OutputSellsy.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputSellsy.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputSellsy.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil, nil,
	)
	goutils.MustBeNil(err)

	searchObjects, err := explorer.ReadObjectsPost(
		api3.AndPathMatcher{
			api3.NewAllowPathStrategy(searchEndpoints),
			api3.NewDenyPathStrategy(ignoreSearchEndpoints),
		},
		searchObjectEndpoints, nil,
		api3.CustomMappingObjectCheck(intercom.ObjectNameToResponseField),
	)
	goutils.MustBeNil(err)

	for _, searchObject := range searchObjects {
		searchObject.ObjectName, _ = strings.CutSuffix(searchObject.ObjectName, "/search")
	}

	// Search objects take precedence
	return readObjects.Combine(searchObjects)
}
