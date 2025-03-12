package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/clickup/metadata"
	"github.com/amp-labs/connectors/providers/clickup/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ //nolint:gochecknoglobals
		"/api/v1/data_connectors",
		"/api/v1/customer_search/show_events",
	}

	objectEndpoints = map[string]string{ //nolint:gochecknoglobals
		"/api/v1/tag_contexts/list": "tag_contexts/list",
		"/api/v1/segments/list":     "segments/list",
	}

	overrideDisplayName = map[string]string{ //nolint:gochecknoglobals
		"tag_contexts/list": "Tag Contexts",
		"segments/list":     "Segments",
		"onsite_slots":      "Onsite Slots",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"email_templates": "results",
		"campaigns":       "results",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	)
)

func main() {
	explorer, err := openapi.ApiManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(api3.CamelCaseToSpaceSeparated, api3.CapitalizeFirstLetterEveryWord),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, overrideDisplayName, api3.CustomMappingObjectCheck(objectNametoResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
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
