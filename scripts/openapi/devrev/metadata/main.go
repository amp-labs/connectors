// Extracts list endpoint schemas from OpenAPI spec and writes providers/devrev/metadata/schemas.json.
package main

import (
	"log"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/devrev/metadata"
	"github.com/amp-labs/connectors/providers/devrev/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// DevRev list endpoints are GET /resource.list; allow only *.list paths.
	allowListPaths = []string{"*.list"}

	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
		"reactions.list": "reactors", // API returns reactors only.
	},
		func(objectName string) string {
			s := strings.TrimSuffix(objectName, ".list")

			return strings.ReplaceAll(s, "-", "_")
		},
	)
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			removeListSuffix, // e.g. CustomerAttributeList -> CustomerAttribute (all objects are collections)
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			removeListResponseSuffix, // e.g. Accounts List Response -> Accounts
			api3.Pluralize,           // e.g. Customer Attribute -> Customer Attributes
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowListPaths),
		nil, nil,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
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

		// Register under object name without .list (e.g. accounts not accounts.list).
		objectName := strings.TrimSuffix(object.ObjectName, ".list")

		for _, field := range object.Fields {
			schemas.Add("", objectName, object.DisplayName, object.URLPath, object.ResponseKey,
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

// removeListSuffix strips the "List" suffix from display names.
func removeListSuffix(displayName string) string {
	s, _ := strings.CutSuffix(displayName, "List")

	return s
}

// removeListResponseSuffix strips " List Response" / " Response" so e.g. "Accounts List Response" -> "Accounts".
func removeListResponseSuffix(displayName string) string {
	if s, ok := strings.CutSuffix(displayName, " List Response"); ok {
		return s
	}

	if s, ok := strings.CutSuffix(displayName, " Response"); ok {
		return s
	}

	return displayName
}
