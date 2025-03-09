package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
	"github.com/amp-labs/connectors/providers/ashby/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ //nolint:gochecknoglobals
		"interviewEvent.list",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	},
		func(objectName string) (fieldName string) {
			return "results"
		},
	)
)

func main() {
	explorer, err := openapi.ApiManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(api3.CamelCaseToSpaceSeparated, api3.CapitalizeFirstLetterEveryWord),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsPost(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil, api3.CustomMappingObjectCheck(objectNametoResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range readObjects { //nolint:gochecknoglobals
		if object.Problem != nil {
			slog.Error("Schema not extracted",
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
