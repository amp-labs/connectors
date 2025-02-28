package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
	"github.com/amp-labs/connectors/providers/brevo/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/smtp/emails",
		"/transactionalSMS/statistics/aggregatedReport",
		"/smtp/statistics/aggregatedReport",
		"/products/batch",
		"/crm/tasktypes",
		"/orders/status/batch",
		"/categories/batch",
		"/contacts/batch",
		"/corporate/masterAccount",
		"/account",
		"/crm/pipeline/details",
		"/smtp/blockedDomains",
	}

	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/companies/attributes":                "companies/attributes",
		"/contacts/attributes":                 "contacts/attributes",
		"/inbound/events":                      "inbound/events",
		"/crm/attributes/deals":                "attributes/deals",
		"/smtp/statistics/reports":             "smtp/statistics/reports",
		"/transactionalSMS/statistics/reports": "transactionalSMS/statistics/reports",
		"/transactionalSMS/statistics/events":  "transactionalSMS/statistics/events",
		"/smtp/statistics/events":              "smtp/statistics/events",
	}

	overrideDisplayName = map[string]string{ // nolint:gochecknoglobals
		"companies/attributes":                "Companies Attributes",
		"contacts/attributes":                 "Contacts Attributes",
		"inbound/events":                      "Inbound Events",
		"crm/attributes/deals":                "Deals Attributes",
		"smtp/statistics/reports":             "SMTP Statistics Reports",
		"transactionalSMS/statistics/reports": "Transactional SMS Statistics Reports",
		"transactionalSMS/statistics/events":  "Transactional SMS Statistics Events",
		"smtp/statistics/events":              "SMTP Statistics Events",
		"attributes/deals":                    "Deals Attributes",
	}

	// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"contacts/attributes":                 "attributes",
		"companies/attributes":                "",
		"inbound/events":                      "events",
		"transactionalSMS/statistics/events":  "events",
		"smtp/statistics/events":              "events",
		"smtp/statistics/reports":             "reports",
		"transactionalSMS/statistics/reports": "reports",
		"blockedContacts":                     "contacts",
		"smsCampaigns":                        "campaigns",
		"subAccount":                          "subAccounts",
		"tasks":                               "items",
		"emailCampaigns":                      "campaigns",
		"companies":                           "items",
		"deals":                               "items",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	)
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
		objectEndpoints, overrideDisplayName, api3.CustomMappingObjectCheck(objectNameToResponseField),
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
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
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
