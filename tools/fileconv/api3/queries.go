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

// ReadObjectsGet is the same as ReadObjectsGet but retrieves schemas for endpoints that perform reading via GET.
// If you need schemas located under GET and POST operations,
// make 2 calls as they will have different arguments in particular PathMatchingStrategy,
// and then Combine two lists of schemas.
func (e Explorer) ReadObjectsGet(
	matchingStrategy PathMatcherStrategy,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) (Schemas, error) {
	return e.ReadObjects("GET", matchingStrategy, objectEndpoints, displayNameOverride, check)
}

// ReadObjectsPost is the same as ReadObjectsGet but retrieves schemas for endpoints that perform reading via POST.
func (e Explorer) ReadObjectsPost(
	matchingStrategy PathMatcherStrategy,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) (Schemas, error) {
	return e.ReadObjects("POST", matchingStrategy, objectEndpoints, displayNameOverride, check)
}

// ReadObjects will explore OpenAPI file returning list of Schemas.
// See every parameter for detailed customization.
//
// operationName - under which REST operation the schema resides. Ex: GET - list reading, POST - search reading.
// pathMatchingStrategy - guides which URL paths to include in search or to ignore. Should be exhaustive list.
// objectEndpoints - URL path mapped to ObjectName.
// Ex: 	/customer/orders -> orders.
//
//	Note: deep connector would need to do the reverse mapping to reconstruct URL given orders objectName.
//
// displayNameOverride - objectName mapped to custom Display name.
// check - callback that returns true if fieldName matched the target location of Object in response.
// Ex: 	if (objectName == orders && fieldName == data) => true
//
//	Given response with fields {meta{}, data{}, pagination{}} for orders object,
//	the implementation indicates that schema will be located under `data`.
func (e Explorer) ReadObjects(
	operationName string,
	matchingStrategy PathMatcherStrategy,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	check ObjectCheck,
) (Schemas, error) {
	schemas := make(Schemas, 0)

	pathMatcher := matchingStrategy.GivePathMatcher()

	for _, path := range e.GetPathItems(pathMatcher, objectEndpoints) {
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
	pathMatcher PathMatcher, endpointResources map[string]string,
) []PathItem {
	items := handy.Map[string, PathItem]{}

	for path, pathObj := range e.schema.GetPaths() {
		if !pathMatcher.IsPathMatching(path) {
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
