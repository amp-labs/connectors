package main

import (
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/brevo"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
	"github.com/amp-labs/connectors/providers/brevo/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/smtp/emails",
		"/transactionalSMS/statistics/aggregatedReport",
		"/smtp/statistics/aggregatedReport",
		"/products/batch",
		"/orders/status/batch",
		"/categories/batch",
		"/contacts/batch",
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
		objectEndpoints, nil, api3.CustomMappingObjectCheck(brevo.ObjectNameToResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	fmt.Println("object lenght", len(readObjects))

	for _, object := range readObjects {
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
