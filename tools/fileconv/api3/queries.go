package api3

import (
	"log/slog"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/getkin/kin-openapi/openapi3"
)

// Explorer allows to traverse schema in most common ways
// relevant for connectors metadata extraction.
type Explorer struct {
	schema *Document
	*parameters
}

// NewExplorer creates explorer on openAPI v3 file.
// See Option to discover how explorer can be customized.
func NewExplorer(data *openapi3.T, opts ...Option) *Explorer {
	return &Explorer{
		schema: &Document{
			delegate: data,
		},
		parameters: createParams(opts),
	}
}

// ReadObjectsGet is the same as ReadObjectsGet but retrieves schemas for endpoints that perform reading via GET.
// If you need schemas located under GET and POST operations,
// make 2 calls as they will have different arguments in particular PathMatchingStrategy,
// and then Combine two lists of schemas.
func (e Explorer) ReadObjectsGet(
	pathMatcher PathMatcher,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	locator ObjectArrayLocator,
) (Schemas, error) {
	return e.ReadObjects("GET", pathMatcher, objectEndpoints, displayNameOverride, locator)
}

// ReadObjectsPost is the same as ReadObjectsGet but retrieves schemas for endpoints that perform reading via POST.
func (e Explorer) ReadObjectsPost(
	pathMatcher PathMatcher,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	locator ObjectArrayLocator,
) (Schemas, error) {
	return e.ReadObjects("POST", pathMatcher, objectEndpoints, displayNameOverride, locator)
}

// ReadObjects will explore OpenAPI file returning list of Schemas.
// See every parameter for detailed customization.
//
// operationName - under which REST operation the schema resides. Ex: GET - list reading, POST - search reading.
// pathMatcher - guides which URL paths to include in search or to ignore.
// objectEndpoints - URL path mapped to ObjectName.
// Ex: 	/customer/orders -> orders.
//
//	Note: deep connector would need to do the reverse mapping to reconstruct URL given orders objectName.
//
// displayNameOverride - objectName mapped to custom Display name.
// locator - callback that returns true if fieldName matched the target location of Object in response.
// Ex: 	if (objectName == orders && fieldName == data) => true
//
//	Given response with fields {meta{}, data{}, pagination{}} for orders object,
//	the implementation indicates that schema will be located under `data`.
func (e Explorer) ReadObjects(
	operationName string,
	pathMatcher PathMatcher,
	objectEndpoints map[string]string,
	displayNameOverride map[string]string,
	locator ObjectArrayLocator,
) (Schemas, error) {
	schemas := make(Schemas, 0)

	for _, path := range e.GetPathItems(AndPathMatcher{
		pathMatcher,
		// There should be no curly brackets no IDs, no nested resources.
		// Read objects are those that have constant string path.
		IDPathIgnorer{},
	}, objectEndpoints) {
		schema, found, err := path.RetrieveSchemaOperation(operationName,
			displayNameOverride, locator,
			e.displayPostProcessing,
			e.operationMethodFilter,
			e.propertyFlattener,
			e.mediaType,
			*e.autoSelectArrayItem,
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
) []*PathItem {
	items := datautils.Map[string, *PathItem]{}
	duplicates := datautils.UniqueLists[string, *PathItem]{}

	for path, pathObj := range e.schema.GetPaths() {
		if !pathMatcher.IsPathMatching(path) {
			// Ignore this endpoint path.
			continue
		}

		objectName, isDuplicate := endpointResources[path]
		if !isDuplicate {
			// ObjectName is empty at this time.
			// We need to do some processing to infer ObjectName from URL path.
			// By default, the last URL part is the ObjectName describing this REST resource.
			parts := strings.Split(path, "/")
			objectName = parts[len(parts)-1]
		}

		item := PathItem{
			objectName: objectName,
			urlPath:    path,
			delegate:   pathObj,
		}

		if oldItem, exists := items[objectName]; exists {
			duplicates.Add(objectName, oldItem) // existing item is a duplicate
			duplicates.Add(objectName, &item)   // new item is considered a duplicate too
		} else {
			// Associate item with object name.
			items[objectName] = &item
		}
	}

	// Report PathItems that use the same object name.
	// They will be excluded from schema extraction and script user will be warned to take action.
	for objectName, pathItems := range duplicates {
		paths := make([]string, len(pathItems))
		for index, item := range pathItems.List() {
			paths[index] = item.urlPath
		}

		slog.Warn("object name is not unique, ignoring",
			"objectName", objectName,
			"collisions", strings.Join(paths, ", "),
		)

		// We have logged each path that shares the same object name.
		// No such object will be included for consistency.
		delete(items, objectName)
	}

	return items.Values()
}
