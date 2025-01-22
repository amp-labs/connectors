package staticschema

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

const RootModuleID common.ModuleID = "root"

// NewMetadata constructs empty Metadata. To populate it with data use Metadata.Add method.
func NewMetadata[F FieldMetadataMap]() *Metadata[F] {
	return &Metadata[F]{
		Modules: make(map[common.ModuleID]Module[F]),
	}
}

// Metadata is data that is saved in the static file.
//
// This structure offers the following main methods:
// Metadata.ObjectNames - registry of object names by modules.
// Metadata.LookupURLPath - resolves URL given object name.
// Metadata.Select - produces output suitable for connectors.ObjectMetadataConnector.
type Metadata[F FieldMetadataMap] struct {
	// Modules is a map of object names to object metadata
	Modules map[common.ModuleID]Module[F] `json:"modules"`
}

type Module[F FieldMetadataMap] struct {
	ID      common.ModuleID      `json:"id"`
	Path    string               `json:"path"`
	Objects map[string]Object[F] `json:"objects"`
}

type Object[F FieldMetadataMap] struct {
	// Provider's display name for the object
	DisplayName string `json:"displayName"`

	// This is the endpoint URLPath to this REST API resource/object.
	// It can be used to make request to "list objects".
	URLPath string `json:"path"`

	// A field name where the object is located within provider response.
	ResponseKey string `json:"responseKey"`

	// Fields is a map of field names to either field display names or extended field metadata.
	Fields F `json:"fields"`

	// DocsURL points to docs endpoint. Optional.
	DocsURL *string `json:"docs,omitempty"`
}

type (
	FieldMetadataMapV1 map[string]string
	FieldMetadataMapV2 map[string]common.FieldMetadata

	FieldMetadataMap interface {
		FieldMetadataMapV1 | FieldMetadataMapV2
	}
)

// All maps must be copies of the original map, which remains immutable.
// The immutable map serves as a unified registry of metadata for the connector.
// ListObjectMetadata returns a copy that consumers are allowed to modify as needed.
func (o Object[F]) getObjectMetadata() *common.ObjectMetadata {
	if fieldsMap, isV1 := (any(o.Fields)).(FieldMetadataMapV1); isV1 {
		return &common.ObjectMetadata{
			DisplayName: o.DisplayName,
			FieldsMap:   datautils.FromMap(fieldsMap).ShallowCopy(),
		}
	}

	fields, isV2 := (any(o.Fields)).(FieldMetadataMapV2)
	if !isV2 {
		// Unknown fieldsMap version.
		return &common.ObjectMetadata{}
	}

	return common.NewObjectMetadata(
		o.DisplayName,
		datautils.FromMap(fields).ShallowCopy(),
	)
}

// Add will appropriately store the data, abiding to data structure rules.
// NOTE: empty module id is treated as root module.
func (m *Metadata[F]) Add( // nolint:funlen
	moduleID common.ModuleID,
	objectName, objectDisplayName, urlPath, responseKey string,
	fieldMetadataMap F,
	docsURL *string,
) {
	moduleID = moduleIdentifier(moduleID)

	module := m.getOrCreateModule(moduleID)

	object, objectExists := module.Objects[objectName]
	defer func() {
		module.Objects[objectName] = object
	}()

	if !objectExists {
		object = Object[F]{
			DisplayName: objectDisplayName,
			URLPath:     urlPath,
			ResponseKey: responseKey,
			Fields:      fieldMetadataMap,
			DocsURL:     docsURL,
		}

		return
	}

	// This code acts as a bridge between a concrete type and a generic type.
	//
	// Key points:
	// - The range operation is not supported for expressions without a core type.
	//   Therefore, the map is type-asserted to its concrete type.
	// - For the concrete type, old and new values are copied into a joined map of that type.
	// - Finally, the map is converted back to the generic type.
	//
	// Note:
	// - This is a workaround for limitations in Go generics.
	// - During compilation, only one execution path is valid for each concrete type.
	//   The other paths are unreachable and conceptually invalid, as they pertain to entirely different types.
	if presentFields, isV1 := (any(object.Fields)).(FieldMetadataMapV1); isV1 {
		fieldsMap := make(FieldMetadataMapV1)
		for k, v := range presentFields {
			fieldsMap[k] = v
		}

		for k, v := range (any(fieldMetadataMap)).(FieldMetadataMapV1) { // nolint:forcetypeassert
			fieldsMap[k] = v
		}

		object.Fields = any(fieldsMap).(F) // nolint:forcetypeassert

		return
	}

	if presentFields, isV2 := (any(object.Fields)).(FieldMetadataMapV2); isV2 {
		fieldsMap := make(FieldMetadataMapV2)
		for k, v := range presentFields {
			fieldsMap[k] = v
		}

		for k, v := range (any(fieldMetadataMap)).(FieldMetadataMapV2) { // nolint:forcetypeassert
			fieldsMap[k] = v
		}

		object.Fields = any(fieldsMap).(F) // nolint:forcetypeassert

		return
	}
}

func (m *Metadata[F]) refactorLongestCommonPath() {
	for moduleID, module := range m.Modules {
		var (
			commonPath string
			isFirst    = true
		)

		for _, object := range module.Objects {
			path := object.URLPath
			if isFirst {
				commonPath = path
				isFirst = false

				continue
			}

			commonPath = commonPrefix(commonPath, path)

			if len(commonPath) == 0 {
				break
			}
		}

		// CommonPath is now found.
		module.withPath(commonPath)
		m.Modules[moduleID] = module
	}
}

func (m *Metadata[F]) getOrCreateModule(moduleID common.ModuleID) Module[F] {
	module, ok := m.Modules[moduleID]
	if !ok {
		// new module
		module = Module[F]{
			ID:      moduleID,
			Path:    "",
			Objects: make(map[string]Object[F]),
		}
		m.Modules[moduleID] = module
	}

	return module
}

// ObjectNames provides a registry of object names grouped by module.
func (m *Metadata[F]) ObjectNames() datautils.UniqueLists[common.ModuleID, string] {
	moduleObjectNames := make(datautils.UniqueLists[common.ModuleID, string])

	for key, value := range m.Modules {
		names := datautils.NewStringSet()
		for name := range value.Objects {
			names.AddOne(name)
		}

		moduleObjectNames[key] = names

		if key == RootModuleID {
			// Empty ModuleID could be passed referring to the same root module.
			moduleObjectNames[""] = names
		}
	}

	return moduleObjectNames
}

// LookupURLPath will give you the URL path for the object located under the module.
// NOTE: empty module id is treated as root module.
func (m *Metadata[F]) LookupURLPath(moduleID common.ModuleID, objectName string) (string, error) {
	moduleID = moduleIdentifier(moduleID)

	path := m.Modules[moduleID].Objects[objectName].URLPath
	if len(path) == 0 {
		return "", common.ErrResolvingURLPathForObject
	}

	fullPath := m.LookupModuleURLPath(moduleID) + path

	return fullPath, nil
}

func (m *Metadata[F]) LookupModuleURLPath(moduleID common.ModuleID) string {
	moduleID = moduleIdentifier(moduleID)

	return m.Modules[moduleID].Path
}

// ModuleRegistry returns the list of API modules from static schema.
func (m *Metadata[F]) ModuleRegistry() common.Modules {
	result := make(common.Modules, len(m.Modules))

	for id, module := range m.Modules {
		// Label and version is not differentiated and all is part of path.
		result[id] = common.Module{
			ID:      module.ID,
			Label:   module.Path,
			Version: "",
		}
	}

	return result
}

// LookupArrayFieldName will give you the field name which holds the array of objects in provider response.
// Ex: CustomerSubscriptions is located under field name subscriptions => { "subscriptions": [{},{},{}] }.
func (m *Metadata[F]) LookupArrayFieldName(moduleID common.ModuleID, objectName string) string {
	moduleID = moduleIdentifier(moduleID)

	fieldName := m.Modules[moduleID].Objects[objectName].ResponseKey

	return fieldName
}

func (m *Module[F]) withPath(path string) {
	// Move last slash from module path to object. It looks better that way.
	path, _ = strings.CutSuffix(path, "/")

	m.Path = path

	// Trim prefix for every object.
	for name, object := range m.Objects {
		object.URLPath, _ = strings.CutPrefix(object.URLPath, path)
		m.Objects[name] = object
	}
}

// In case an empty ModuleID is provided we fall back to the default root module id.
func moduleIdentifier(id common.ModuleID) common.ModuleID {
	if len(id) == 0 {
		return RootModuleID
	}

	return id
}

func commonPrefix(a, b string) string {
	first := []byte(a)
	second := []byte(b)
	shortestLength := len(a)

	if len(a) > len(b) {
		first = []byte(b)
		second = []byte(a)
		shortestLength = len(b)
	}

	result := ""

	for i := range shortestLength {
		if first[i] != second[i] {
			return result
		}

		result += string(first[i])
	}

	return result
}
