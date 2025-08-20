package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/services/reports/status",
		"/services/bulk/job/status",
		"/services/core/model",
		"/services/core/session/id",
		"/objects/core/system-view",
		"/services/construction-forecasting/wip-journal/entry-history",
	}

	objectEndpoints = map[string]string{ //nolint:gochecknoglobals
		"/services/company-config/dimensions/list":                  "dimensions",
		"/objects/accounts-payable/account-label-tax-groups":        "account-label-tax-group",
		"/objects/accounts-payable/account-label":                   "accounts-payable-account-label",
		"/objects/accounts-payable/adjustment":                      "accounts-payable-adjustment",
		"/objects/accounts-payable/adjustment-line":                 "accounts-payable-adjustment-line",
		"/objects/accounts-payable/adjustment-tax-entry":            "accounts-payable-adjustment-tax-entry",
		"/objects/accounts-payable/advance":                         "accounts-payable-advance",
		"/objects/accounts-payable/advance-line":                    "accounts-payable-advance-line",
		"/objects/accounts-payable/payment":                         "accounts-payable-payment",
		"/objects/accounts-receivable/payment-detail":               "accounts-receivable-payment-detail",
		"/objects/accounts-receivable/payment-line":                 "accounts-receivable-payment-line",
		"/objects/accounts-receivable/summary":                      "accounts-receivable-summary",
		"/objects/accounts-receivable/term":                         "accounts-receivable-term",
		"/objects/order-entry/txn-definition":                       "order-entry-txn-definition",
		"/objects/order-entry/subtotal-template-line":               "order-entry-subtotal-template-line",
		"/objects/purchasing/txn-definition-additional-gl-detail":   "purchasing-txn-definition-additional-gl-detail",
		"/objects/purchasing/txn-definition":                        "purchasing-txn-definition",
		"/objects/purchasing/subtotal-template-line":                "purchasing-subtotal-template-line",
		"/objects/purchasing/subtotal-template":                     "purchasing-subtotal-template",
		"/objects/purchasing/recurring-document-subtotal":           "purchasing-recurring-document-subtotal",
		"/objects/purchasing/recurring-document-line":               "purchasing-recurring-document-line",
		"/objects/purchasing/recurring-document":                    "purchasing-recurring-document",
		"/objects/accounts-receivable/adjustment-tax-entry":         "accounts-receivable-adjustment-tax-entry",
		"/objects/accounts-receivable/advance":                      "accounts-receivable-advance",
		"/objects/accounts-receivable/adjustment-line":              "accounts-receivable-adjustment-line",
		"/objects/accounts-receivable/adjustment":                   "accounts-receivable-adjustment",
		"/objects/accounts-receivable/account-label-tax-group":      "accounts-receivable-account-label-tax-group",
		"/objects/inventory-control/document-line":                  "inventory-control-document-line",
		"/objects/inventory-control/document-history":               "inventory-control-document-history",
		"/objects/inventory-control/price-list":                     "inventory-control-price-list",
		"/objects/inventory-control/txn-definition-cogs-gl-detail":  "inventory-control-txn-definition-cogs-gl-detail",
		"/objects/inventory-control/txn-definition-subtotal-detail": "inventory-control-txn-definition-subtotal-detail",
		"/objects/order-entry/document":                             "order-entry-document",
		"/objects/order-entry/document-history":                     "order-entry-document-history",
		"/objects/order-entry/document-line":                        "order-entry-document-line",
		"/objects/order-entry/document-line-detail":                 "order-entry-document-line-detail",
		"/objects/order-entry/document-line-subtotal":               "order-entry-document-line-subtotal",
	}

	ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"v1/notes/": "results",
	},
		func(objectName string) (fieldName string) {
			return "ia::result"
		})
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil, api3.CustomMappingObjectCheck(ObjectNameToResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range readObjects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, api3.CapitalizeFirstLetterEveryWord(api3.KebabCaseToSpaceSeparated(object.ObjectName)), object.URLPath, object.ResponseKey, //nolint:lll
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
