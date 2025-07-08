package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/avoma/metadata"
	"github.com/amp-labs/connectors/providers/avoma/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/v1/calls/{external_id}/",
		"/v1/custom_categories/{uuid}/",
		"/v1/meeting_segments/",
		"/v1/meeting_sentiments/",
		"/v1/meetings/{meeting_uuid}/insights/",
		"/v1/meetings/{uuid}/",
		"/v1/meetings/{uuid}/drop/",
		"/v1/recordings/",
		"/v1/recordings/{uuid}/",
		"/v1/scorecards/{uuid}/",
		"/v1/smart_categories/{uuid}/",
		"/v1/template/{uuid}/",
		"/v1/transcriptions/{uuid}/",
		"/v1/users/{uuid}/",
	}

	ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"v1/notes/":                 "results",
		"v1/scorecard_evaluations/": "results",
		"v1/meetings/":              "results",
		"v1/custom_categories/":     "results",
		"v1/calls/":                 "results",
		"v1/smart_categories/":      "results",
		"v1/transcriptions/":        "results",
	},
		func(objectName string) (fieldName string) {
			return objectName
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
		nil, nil, api3.CustomMappingObjectCheck(ObjectNameToResponseField),
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
		}

		for _, field := range object.Fields {
			objName := strings.TrimPrefix(object.ObjectName, "v1/")
			objName = strings.TrimSuffix(objName, "/")
			displayName := titleCaser.String(objName)
			urlPath := strings.TrimPrefix(object.URLPath, "/v1")

			schemas.Add("", objName, displayName, urlPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
