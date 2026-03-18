// Extracts list endpoint schemas from OpenAPI spec and writes providers/greenhouse/metadata/schemas.json.
package main

import (
	"log"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/greenhouse/metadata"
	"github.com/amp-labs/connectors/providers/greenhouse/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// nolint:gochecknoglobals
var (
	// Greenhouse Harvest API v3 list endpoints use GET /v3/{resource}.
	// Only paths under /v3/ are relevant; ignore auth endpoints.
	ignoreEndpoints = []string{
		"/auth/token",
	}

	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{}, //nolint:gochecknoglobals
		func(objectName string) string {
			// Greenhouse v3 list endpoints return bare JSON arrays (no wrapper key).
			return ""
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

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	titleCaser := cases.Title(language.English)

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		objectName := strings.TrimPrefix(object.ObjectName, "v3/")
		displayName := titleCaser.String(strings.ReplaceAll(objectName, "_", " "))

		for _, field := range object.Fields {
			schemas.Add("", objectName, displayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}
