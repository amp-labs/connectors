package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"strings"

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

	ignoreEndpoints = []string{}
)

// TopLevelPathMatcher only allows paths that have {project_id} and no other ID parameters
// This filters out nested routes like /apps/{app_id}/... or /entitlements/{entitlement_id}/...
// or /customers/{customer_id}/...
type TopLevelPathMatcher struct{}

func (TopLevelPathMatcher) IsPathMatching(path string) bool {
	// Count ID parameters in the path
	idCount := 0
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			idCount++
			// Only allow {project_id}, reject all other IDs
			if part != "{project_id}" {
				return false
			}
		}
	}

	// Must have exactly one ID parameter which is {project_id}
	return idCount == 1
}

// extractEndpoints extracts all top-level endpoints and creates objectEndpoints mapping
func extractEndpoints(explorer *api3.Explorer[any]) (map[string]string, error) {
	// Get endpoint operations to extract paths
	endpoints, err := explorer.GetEndpointOperations(TopLevelPathMatcher{}, http.MethodGet)
	if err != nil {
		return nil, err
	}

	objectEndpoints := make(map[string]string)

	for _, endpoint := range endpoints {
		urlPath := endpoint.URLPath
		// Extract object name by removing /projects/{project_id}/ prefix
		objectName, found := strings.CutPrefix(urlPath, "/projects/{project_id}/")
		if !found {
			continue
		}
		if objectName == "" {
			continue
		}

		// Map the full URL path to the object name
		objectEndpoints[urlPath] = objectName
	}

	return objectEndpoints, nil
}

// removeListSuffix removes "List" suffix from display names (case-insensitive)
func removeListSuffix(displayName string) string {
	displayName = strings.TrimSuffix(displayName, "List")
	displayName = strings.TrimSuffix(displayName, "list")
	displayName = strings.TrimSuffix(displayName, "LIST")
	return displayName
}

func main() {
	explorer, err := FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			removeListSuffix,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			api3.Pluralize,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	// Step 1: Extract endpoints first
	objectEndpoints, err := extractEndpoints(explorer)
	goutils.MustBeNil(err)

	displayNameOverride := map[string]string{}

	// Step 2: Read objects using the extracted endpoints mapping
	objects, err := explorer.ReadObjects(
		http.MethodGet,
		TopLevelPathMatcher{},
		objectEndpoints,
		displayNameOverride,
		// RevenueCat responses use "items" field for arrays
		func(objectName, fieldName string) bool {
			return fieldName == "items"
		},
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	// Step 3: Process objects
	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		// Extract object name from URL path
		// object.ObjectName should already be set from objectEndpoints mapping
		objectName := object.ObjectName

		// The path should keep /{project_id}/ prefix for URL construction
		// Transform /projects/{project_id}/... to /{project_id}/...
		urlPath, _ := strings.CutPrefix(object.URLPath, "/projects")
		if !strings.HasPrefix(urlPath, "/") {
			urlPath = "/" + urlPath
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, objectName, object.DisplayName, urlPath, object.ResponseKey,
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
