package main

import (
	"log/slog"
	"regexp"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/openapi/atlassian/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Singular objects.
		"/admin-key",
		"/data-policies/metadata",
		// Requires query parameters.
		"/custom-content",
	}
)

func main() {
	explorer, err := files.InputConfluenceV2.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			extractNestedName,
			api3.Pluralize,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil,
		func(objectName, fieldName string) bool {
			// The array of records is located under "results" key.
			return fieldName == "results"
		},
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(providers.ModuleAtlassianConfluence,
				object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(files.OutputConfluenceV2.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputConfluenceV2.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

// The object name may be located between triangular brackets.
// Examples:
//
//	MultiEntityResult<SpacePermission>
//	MultiEntityResult<BlogPost>
//	MultiEntityResult<CustomContent>
//	MultiEntityResult<InlineCommentModel>
func extractNestedName(displayName string) string {
	re := regexp.MustCompile(`<([^<>]+)>`)
	match := re.FindStringSubmatch(displayName)

	if len(match) == 2 { // nolint:mnd
		return match[1]
	}

	return displayName
}
