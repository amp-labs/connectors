// Extracts list endpoint schemas from OpenAPI spec and writes providers/housecallPro/metadata/schemas.json.
package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/housecallPro/metadata"
	"github.com/amp-labs/connectors/providers/housecallPro/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

//
//nolint:gochecknoglobals
var ignoreEndpoints = []string{
	// endpoints that are not collections
	"/application",
	"/company/schedule_availability",
	"/company",
}

// Most Housecall Pro list routes omit the "/api" path prefix;
// only price book list endpoints use /api/price_book/...
//
//nolint:gochecknoglobals
var objectEndpoints = map[string]string{
	"/api/price_book/material_categories": "price_book/material_categories",
	"/api/price_book/materials":           "price_book/materials",
	"/api/price_book/price_forms":         "price_book/price_forms",
	"/api/price_book/services":            "price_book/services",
	"/job_fields/job_types":               "job_fields/job_types",
}

//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	"price_book/material_categories": "Material Categories",
	"price_book/materials":           "Materials",
	"price_book/services":            "Price Book Services",
	"price_book/price_forms":         "Price Forms",
	"job_fields/job_types":           "Job Types",
}

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			api3.Pluralize,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverrides, api3.DataObjectLocator,
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

			continue
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

	log.Println("Completed.")
}
