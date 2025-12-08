package staticschema

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// NewMetadata constructs empty Metadata. To populate it with data use Metadata.Add method.
func NewMetadata[F FieldMetadataMap]() *Metadata[F, any] {
	return NewExtendedMetadata[F, any]()
}

// NewExtendedMetadata constructs empty Metadata. To populate it with data use Metadata.Add method.
// Metadata may include custom properties describing object schema.
func NewExtendedMetadata[F FieldMetadataMap, C any]() *Metadata[F, C] {
	return &Metadata[F, C]{
		Modules: make(map[common.ModuleID]Module[F, C]),
	}
}

// Metadata is data that is saved in the static file.
//
// This structure offers the following main methods:
// Metadata.ObjectNames - registry of object names by modules.
// Metadata.LookupURLPath - resolves URL given object name.
// Metadata.Select - produces output suitable for connectors.ObjectMetadataConnector.
type Metadata[F FieldMetadataMap, C any] struct {
	// Modules is a map of object names to object metadata
	Modules map[common.ModuleID]Module[F, C] `json:"modules"`
}

type Module[F FieldMetadataMap, C any] struct {
	ID      common.ModuleID         `json:"id"`
	Path    string                  `json:"path"`
	Objects map[string]Object[F, C] `json:"objects"`
}

// Object describes provider resource.
// Fields of an object supports multiple formats.
// Custom properties can be associated with each object.
type Object[F FieldMetadataMap, C any] struct {
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

	// Custom includes properties important for altering the behaviour of Read/ListObjectMetadata.
	// This is relevant for some connectors.
	Custom C `json:"custom,omitempty"`
}

type FieldMetadata struct {
	DisplayName  string           `json:"displayName,omitempty"`
	ValueType    common.ValueType `json:"valueType,omitempty"`
	ProviderType string           `json:"providerType,omitempty"`
	ReadOnly     *bool            `json:"readOnly,omitempty"`
	Values       FieldValues      `json:"values,omitempty"`
}

type FieldValues []FieldValue

type FieldValue struct {
	Value        string `json:"value,omitempty"`
	DisplayValue string `json:"displayValue,omitempty"`
}

type (
	FieldMetadataMapV1 map[string]string
	FieldMetadataMapV2 map[string]FieldMetadata

	FieldMetadataMap interface {
		FieldMetadataMapV1 | FieldMetadataMapV2
	}
)

func (m FieldMetadataMapV2) convertToCommon() map[string]common.FieldMetadata {
	result := make(map[string]common.FieldMetadata)

	for fieldName, field := range m {
		values := make(common.FieldValues, len(field.Values))
		for index, value := range field.Values {
			values[index] = common.FieldValue{
				Value:        value.Value,
				DisplayValue: value.DisplayValue,
			}
		}

		if len(values) == 0 {
			values = nil
		}

		result[fieldName] = common.FieldMetadata{
			DisplayName:  field.DisplayName,
			ValueType:    field.ValueType,
			ProviderType: field.ProviderType,
			ReadOnly:     field.ReadOnly,
			Values:       values,
		}
	}

	return result
}

// All maps must be copies of the original map, which remains immutable.
// The immutable map serves as a unified registry of metadata for the connector.
// ListObjectMetadata returns a copy that consumers are allowed to modify as needed.
func (o Object[F, C]) getObjectMetadata() *common.ObjectMetadata {
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
		fields.convertToCommon(),
	)
}

// Add will appropriately store the data, abiding to data structure rules.
// NOTE: empty module id is treated as root module.
func (m *Metadata[F, C]) Add( // nolint:funlen
	moduleID common.ModuleID,
	objectName, objectDisplayName, urlPath, responseKey string,
	fieldMetadataMap F, docsURL *string,
	customProperties C,
) {
	moduleID = moduleIdentifier(moduleID)

	module := m.getOrCreateModule(moduleID)

	object, objectExists := module.Objects[objectName]
	defer func() {
		module.Objects[objectName] = object
	}()

	if !objectExists {
		object = Object[F, C]{
			DisplayName: objectDisplayName,
			URLPath:     urlPath,
			ResponseKey: responseKey,
			Fields:      fieldMetadataMap,
			DocsURL:     docsURL,
			Custom:      customProperties,
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

func (m *Metadata[F, C]) refactorLongestCommonPath() {
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

func (m *Metadata[F, C]) getOrCreateModule(moduleID common.ModuleID) Module[F, C] {
	module, ok := m.Modules[moduleID]
	if !ok {
		// new module
		module = Module[F, C]{
			ID:      moduleID,
			Path:    "",
			Objects: make(map[string]Object[F, C]),
		}
		m.Modules[moduleID] = module
	}

	return module
}

// ObjectNames provides a registry of object names grouped by module.
func (m *Metadata[F, C]) ObjectNames() datautils.UniqueLists[common.ModuleID, string] {
	moduleObjectNames := make(datautils.UniqueLists[common.ModuleID, string])

	for key, value := range m.Modules {
		names := datautils.NewStringSet()
		for name := range value.Objects {
			names.AddOne(name)
		}

		moduleObjectNames[key] = names

		if key == common.ModuleRoot {
			// Empty ModuleID could be passed referring to the same root module.
			moduleObjectNames[""] = names
		}
	}

	return moduleObjectNames
}

// LookupURLPath will give you the URL path for the object located under the module.
// NOTE: empty module id is treated as root module.
//
// Deprecated.
// Use FindURLPath. The module path will be removed from static files.
func (m *Metadata[F, C]) LookupURLPath(moduleID common.ModuleID, objectName string) (string, error) {
	path, err := m.FindURLPath(moduleID, objectName)
	if err != nil {
		return "", err
	}

	moduleID = moduleIdentifier(moduleID)
	fullPath := m.LookupModuleURLPath(moduleID) + path

	return fullPath, nil
}

func (m *Metadata[F, C]) FindURLPath(moduleID common.ModuleID, objectName string) (string, error) {
	moduleID = moduleIdentifier(moduleID)

	path := m.Modules[moduleID].Objects[objectName].URLPath
	if len(path) == 0 {
		return "", common.ErrResolvingURLPathForObject
	}

	return path, nil
}

func (m *Metadata[F, C]) LookupModuleURLPath(moduleID common.ModuleID) string {
	moduleID = moduleIdentifier(moduleID)

	return m.Modules[moduleID].Path
}

// ModuleRegistry returns the list of API modules from static schema.
func (m *Metadata[F, C]) ModuleRegistry() common.Modules {
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
func (m *Metadata[F, C]) LookupArrayFieldName(moduleID common.ModuleID, objectName string) string {
	fieldName, _ := m.FindArrayFieldName(moduleID, objectName)

	return fieldName
}

func (m *Metadata[F, C]) FindArrayFieldName(moduleID common.ModuleID, objectName string) (string, bool) {
	moduleID = moduleIdentifier(moduleID)

	module, ok := m.Modules[moduleID]
	if !ok {
		return "", false
	}

	object, ok := module.Objects[objectName]
	if !ok {
		return "", false
	}

	fieldName := object.ResponseKey

	return fieldName, true
}

func (m *Module[F, C]) withPath(path string) {
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
		return common.ModuleRoot
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
