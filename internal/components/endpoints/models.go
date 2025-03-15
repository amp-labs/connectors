package endpoints

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

const RecordIDKey = "{{.recordID}}"

// OperationRegistry is a comprehensive blueprint that organizes a collection of OperationSpec instances.
// It is structured by module ID and object name.
type OperationRegistry map[common.ModuleID]datautils.DefaultMap[string, OperationSpec]

// NewOperationRegistry constructs a new OperationRegistry. It sets the default HTTP method for each operation
// unless an override is provided in the registry map. This function simplifies the creation of an
// OperationRegistry by automatically handling default values.
func NewOperationRegistry(
	defaultHTTPMethod string,
	registry map[common.ModuleID]map[string]OperationSpec,
	fallback func(common.ModuleID, string) OperationSpec,
) OperationRegistry {
	result := make(map[common.ModuleID]datautils.DefaultMap[string, OperationSpec])

	for moduleID, mapping := range registry {
		// OperationSpec should have default value if none is specified.
		// Usually most write operations have identical operations,
		// this makes the registry shorter, drawing attention to exceptions.
		for objectName, spec := range mapping {
			if len(spec.Method) == 0 {
				spec.Method = defaultHTTPMethod
				mapping[objectName] = spec
			}
		}

		result[moduleID] = datautils.NewDefaultMap(mapping,
			createFallback(fallback, moduleID),
		)
	}

	return result
}

func createFallback(
	fallback func(common.ModuleID, string) OperationSpec,
	moduleID common.ModuleID,
) func(objectName string) OperationSpec {
	if fallback == nil {
		return func(objectName string) OperationSpec {
			return OperationSpec{}
		}
	}

	return func(objectName string) OperationSpec {
		return fallback(moduleID, objectName)
	}
}

func (r OperationRegistry) ObjectNames() datautils.UniqueLists[common.ModuleID, string] {
	moduleObjectNames := make(datautils.UniqueLists[common.ModuleID, string])

	for moduleID, module := range r {
		objects := module.Keys()
		moduleObjectNames.Add(moduleID, objects...)
	}

	return moduleObjectNames
}

// OperationSpec serves as a template for defining API calls.
// This specification will act as a prototype for constructing an OperationContext at runtime.
//
// It includes properties that differentiate endpoints, such as:
//   - Method: The HTTP method to use for the operation.
//     For example, most update operations use PUT, while some may use PATCH.
//   - Path: A required field that defines the relative unique URL path for each object.
//
// TODO: Consider expanding this struct to include custom payload construction, headers, and other configurations.
type OperationSpec struct {
	Method string
	Path   string
}

func (d OperationSpec) isEmpty() bool {
	return len(d.Method) == 0 && len(d.Path) == 0
}

func (d OperationSpec) getURLPath(recordID string) string {
	if len(recordID) == 0 {
		// Usually this is a create or command endpoint.
		return d.Path
	}

	// No template. Usually record identifier is attached at the end of endpoint.
	if !strings.Contains(d.Path, RecordIDKey) {
		return d.Path + "/" + recordID
	}

	// Insert recordID inside URL according to the template format.
	return strings.ReplaceAll(d.Path, RecordIDKey, recordID)
}

// ResponseIdentifierRegistry is a collection of JSON paths to locate record identifiers in the response body.
// It is structured by module ID and object name.
type ResponseIdentifierRegistry map[common.ModuleID]datautils.DefaultMap[string, JSONPath]

// JSONPath defines a data location using a key and optional nested JSON keys.
type JSONPath struct {
	key    string
	nested []string
}

func NewJSONPath(key string, nested ...string) JSONPath {
	return JSONPath{
		key:    key,
		nested: nested,
	}
}
