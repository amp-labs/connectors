package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/openapi/clio/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Endpoint that returns singleton/system data rather than a listable collection.
var ignoreEndpoints = []string{ // nolint:gochecknoglobals
	"/users/who_am_i",
}

func main() {
	explorer, err := files.InputGrow.GetExplorer(
		api3.WithMediaType("application/json; charset=utf-8"),
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil,
		api3.DataObjectLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}
	titleCaser := cases.Title(language.English)

	for _, object := range readObjects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		objectName := normalizeObjectName(object.ObjectName)
		displayName := titleCaser.String(strings.ReplaceAll(objectName, "_", " "))

		for _, field := range object.Fields {
			schemas.Add(providers.ModuleClioGrow, objectName, displayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	goutils.MustBeNil(files.OutputGrow.SaveSchemas(schemas))
	goutils.MustBeNil(files.OutputGrow.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func normalizeObjectName(objectName string) string {
	objectName = strings.TrimPrefix(objectName, "/")
	objectName = strings.TrimSuffix(objectName, ".json")
	objectName = strings.TrimSuffix(objectName, "/index")

	return objectName
}
