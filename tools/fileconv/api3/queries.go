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

// GetBasicReadObjects retrieves schemas that can be used by ListObjectMetadata.
// objectEndpoints - optional map of Endpoint Path to ObjectName associated with it.
// ignoreEndpoints - optional list of paths to ignore
// displayNameOverride - optional map of ObjectName to DisplayName. This will override display name from OpenAPI doc.
func (e Explorer) GetBasicReadObjects(
	ignoreEndpoints []string,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) ([]Schema, error) {
	schemas := make([]Schema, 0)

	for _, path := range e.GetBasicPathItems(newIgnorePathStrategy(ignoreEndpoints), objectEndpoints) {
		schema, found, err := path.RetrieveSchemaOperationGet(displayNameOverride, check, e.displayPostProcessing)
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

// GetBasicPathItems returns path items where object name is a single word.
func (e Explorer) GetBasicPathItems(
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
