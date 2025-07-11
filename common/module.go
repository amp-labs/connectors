package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	// ErrInvalidModuleDeclaration is returned when a module entry in the Modules map is incorrectly defined.
	// This may occur if the map key does not match the module's declared ID,
	// or if required provider/module metadata is missing or malformed.
	ErrInvalidModuleDeclaration = errors.New("supported modules are not correctly defined")

	// ErrMissingModule can be returned when connector cannot resolve ModuleID.
	ErrMissingModule = errors.New("module not found")

	// ErrUnsupportedModule returned when provided module is not supported.
	ErrUnsupportedModule = errors.New("provided module is not supported")
)

const ModuleRoot ModuleID = "root"

type ModuleID string

type Modules = datautils.Map[ModuleID, Module]

// Module represents a set of endpoints and functionality available by provider.
// Single provider may support multiple modules, requiring the user to choose a module before making requests.
// Modules may differ by version or by theme, covering different systems or functionalities.
type Module struct {
	ID      ModuleID
	Label   string // e.g. "crm"
	Version string // e.g. "v3"
}

func (a Module) Path() string {
	if len(a.Label) == 0 {
		return a.Version
	}

	return fmt.Sprintf("%s/%s", a.Label, a.Version)
}

// ModuleObjectNameToFieldName is a grouping of ObjectName to response field name mappings defined for each Module.
//
// Explanation: modules have objects, each object is located under certain field name in the response body.
// This mapping stores the general relationship between the said ObjectName and FieldName.
// Those objects that do not follow the pattern described in the fallback method are hard code as exceptions.
//
// Ex:
//
//	Given:	Connector has 2 modules -- Commerce, Messaging.
//			Commerce module has objects stored under "data" field name, except "carts".
//			Messaging module has objects stored under the same name as object, except "chats".
//	Then:	It will be represented as follows:
//
//		ModuleObjectNameToFieldName{
//			ModuleCommerce: datautils.NewDefaultMap(map[string]string{
//				"carts": "carts",
//			},
//				func(objectName string) string {
//					return "data" // always under "data" field {"data": [{},{},...]}
//				},
//			),
//			ModuleHelpCenter: datautils.NewDefaultMap(map[string]string{
//				"chats":        "active_chats",
//			}, func(objectName string) string {
//				fieldName := objectName // Object "messages" is stored under {"messages": [{},{},...]}
//				return fieldName
//			}),
//		}
type ModuleObjectNameToFieldName map[ModuleID]datautils.DefaultMap[string, string]

type OperationDescription struct {
	Operation string
	Path      string
}

func (d OperationDescription) IsEmpty() bool {
	return len(d.Operation) == 0 && len(d.Path) == 0
}

func (d OperationDescription) GetURLPath(recordID string) string {
	if len(recordID) == 0 {
		// Usually this is a create or command endpoint.
		return d.Path
	}

	// No template. Usually record identifier is attached at the end of endpoint.
	if !strings.Contains(d.Path, "{{.recordID}}") {
		return d.Path + "/" + recordID
	}

	// Insert recordID inside URL according to the template format.
	return strings.ReplaceAll(d.Path, "{{.recordID}}", recordID)
}

type ModuleObjectNameToOperationDescription map[ModuleID]datautils.DefaultMap[string, OperationDescription]

func NewModuleObjectNameToOperationDescription(
	defaultHTTPMethod string,
	registry map[ModuleID]map[string]OperationDescription,
) ModuleObjectNameToOperationDescription {
	result := make(map[ModuleID]datautils.DefaultMap[string, OperationDescription])

	for moduleID, mapping := range registry {

		// OperationDescription should have default value if none is specified.
		// Usually most write operations have identical operations,
		// this makes the registry shorter, drawing attention to exceptions.
		for objectName, description := range mapping {
			if len(description.Operation) == 0 {
				description.Operation = defaultHTTPMethod
				mapping[objectName] = description
			}
		}

		result[moduleID] = datautils.NewDefaultMap(mapping,
			func(name string) OperationDescription {
				return OperationDescription{}
			},
		)
	}

	return result
}
