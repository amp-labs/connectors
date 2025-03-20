package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/gitlab/metadata"
	"github.com/amp-labs/connectors/providers/gitlab/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ //nolint:gochecknoglobals
		"/api/v4/projects/import",
		"/api/v4/user/personal_access_tokens",
		"/api/v4/container_registry_event/events",
		"/api/v4/integrations/slack/events",
		"/api/v4/user/status",
		"/api/v4/runners/all",
		"/api/v4/groups/import/authorize",
		"/api/v4/projects/import-relation/authorize",
		"/api/v4/projects/import/authorize",
		"/api/v4/keys",
		"/api/v4/user/runners",
		"/api/v4/user/avatar",
		"/api/v4/avatar",
	}

	objectEndpoints = map[string]string{ //nolint:gochecknoglobals
		"/user/issues": "user/issues",
	}

	overrideDisplayName = map[string]string{ //nolint:gochecknoglobals
		"search/issues": "Search Issues",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"user/installations":        "installations",
		"installation/repositories": "repositories",
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
