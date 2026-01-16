package main

import (
	_ "embed"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/revenuecat/metadata"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	//go:embed swagger.yaml
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{}
)

func main() {
	explorer, err := FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CapitalizeFirstLetterEveryWord,
		),
		// api3.WithMediaType("application/yaml"),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	// Use ReadObjects instead of ReadObjectsGet to allow paths with IDs
	// RevenueCat uses paths like /projects/{project_id}/customers for list endpoints
	objects, err := explorer.ReadObjects(
		http.MethodGet,
		api3.DefaultPathMatcher{},
		nil, nil,
		// RevenueCat responses use "items" field for arrays
		func(objectName, fieldName string) bool {
			return fieldName == "items"
		},
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted", "objectName", object.ObjectName, "error", object.Problem)
		}

		// Use object name directly from OpenAPI parser (from YAML)
		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
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
