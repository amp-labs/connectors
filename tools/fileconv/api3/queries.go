package api3

import (
	"log/slog"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common/handy"
)

// Explorer allows to traverse schema in most common ways
// relevant for connectors metadata extraction.
type Explorer struct {
	schema *Document
	*parameters
}

// ReadObjectsGet retrieves schemas that can be used by ListObjectMetadata.
// objectEndpoints - optional map of Endpoint Path to ObjectName associated with it.
// ignoreEndpoints - optional list of paths to ignore
// displayNameOverride - optional map of ObjectName to DisplayName. This will override display name from OpenAPI doc.
func (e Explorer) ReadObjectsGet(
	ignoreEndpoints []string,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) (Schemas, error) {
	return e.ReadObjects("GET", ignoreEndpoints, true, objectEndpoints, displayNameOverride, check)
}

// SearchObjectsPost is the same as ReadObjectsGet but retrieves schemas for endpoints that perform reading via POST.
func (e Explorer) SearchObjectsPost(
	fromEndpoints []string,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) (Schemas, error) {
	return e.ReadObjects("POST", fromEndpoints, false, objectEndpoints, displayNameOverride, check)
}

// ReadObjects has 2 reading modes.
// One ignores the provided list of endpoints (ignore=true).
// The other scopes the search to just the list of endpoints (ignore=false).
func (e Explorer) ReadObjects(
	operationName string,
	endpoints []string,
	ignore bool,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) (Schemas, error) {
	schemas := make(Schemas, 0)

	for _, path := range e.GetPathItems(newIgnorePathStrategy(endpoints, ignore), objectEndpoints) {
		schema, found, err := path.RetrieveSchemaOperation(operationName,
			displayNameOverride, check, e.displayPostProcessing, e.parameterFilter,
		)
		if err != nil {
			return nil, err
		}

		if found {
			// schema was found save it
			schemas = append(schemas, *schema)
		}
	}

	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Problem == nil && schemas[j].Problem != nil
	})

	return schemas, nil
}

// GetPathItems returns path items where object name is a single word.
func (e Explorer) GetPathItems(
	ignoreEndpoints *ignorePathStrategy, endpointResources map[string]string,
) []PathItem {
	items := handy.Map[string, PathItem]{}

	for path, pathObj := range e.schema.GetPaths() {
		if ignoreEndpoints.Check(path) {
			// Ignore this endpoint path.
			continue
		}

		if strings.Contains(path, "{") {
			// as of now only single word objects are supported
			// there should be no slashes, curly brackets - nested resources
			continue
		}

		objectName, ok := endpointResources[path]
		if !ok {
			// ObjectName is empty at this time.
			// We need to do some processing to infer ObjectName from URL path.
			// By default, the last URL part is the ObjectName describing this REST resource.
			parts := strings.Split(path, "/")
			objectName = parts[len(parts)-1]
		}

		if items.Has(objectName) {
			slog.Warn("object name is not unique, ignoring",
				"objectName", objectName,
				"path", path,
				"collidesWith", items[objectName].urlPath,
			)
		}

		items[objectName] = PathItem{
			objectName: objectName,
			urlPath:    path,
			delegate:   pathObj,
		}
	}

	return items.Values()
}
