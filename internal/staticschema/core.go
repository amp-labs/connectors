package staticschema

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

const RootModuleID common.ModuleID = "root"

// NewMetadata constructs empty Metadata. To populate it with data use Metadata.Add method.
func NewMetadata[F ObjectFields]() *Metadata[F] {
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
type Metadata[F ObjectFields] struct {
	// Modules is a map of object names to object metadata
	Modules map[common.ModuleID]Module[F] `json:"modules"`
}

type Module[F ObjectFields] struct {
	ID      common.ModuleID      `json:"id"`
	Path    string               `json:"path"`
	Objects map[string]Object[F] `json:"objects"`
}

type Object[F ObjectFields] struct {
	// Provider's display name for the object
	DisplayName string `json:"displayName"`

	// This is the endpoint URLPath to this REST API resource/object.
	// It can be used to make request to "list objects".
	URLPath string `json:"path"`

	// A field name where the object is located within provider response.
	ResponseKey string `json:"responseKey"`

	// FieldsMap is a map of field names to field display names
	FieldsMap F `json:"fields"`

	// DocsURL points to docs endpoint. Optional.
	DocsURL *string `json:"docs,omitempty"`
}

type (
	ObjectFieldsV1 map[string]string
	ObjectFieldsV2 map[string]common.FieldMetadata

	ObjectFields interface {
		ObjectFieldsV1 | ObjectFieldsV2
	}
)

func (v Object[F]) getObjectMetadata() *common.ObjectMetadata {
	if fieldsMap, isV1 := (any(v.FieldsMap)).(ObjectFieldsV1); isV1 {
		return &common.ObjectMetadata{
			DisplayName: v.DisplayName,
			FieldsMap:   datautils.FromMap(fieldsMap).ShallowCopy(),
		}
	}

	fields, isV2 := (any(v.FieldsMap)).(ObjectFieldsV2)
	if !isV2 {
		// Unknown fieldsMap version.
		return &common.ObjectMetadata{}
	}

	return common.NewObjectMetadata(
		v.DisplayName,
		datautils.FromMap(fields).ShallowCopy(),
	)
}

// Add will appropriately store the data, abiding to data structure rules.
// NOTE: empty module id is treated as root module.
func (r *Metadata[F]) Add(
	moduleID common.ModuleID,
	objectName, objectDisplayName, urlPath, responseKey string,
	records F,
	docsURL *string,
) {
	moduleID = moduleIdentifier(moduleID)

	module := r.getOrCreateModule(moduleID)

	object, objectExists := module.Objects[objectName]
	defer func() {
		module.Objects[objectName] = object
	}()

	if !objectExists {
		object = Object[F]{
			DisplayName: objectDisplayName,
			URLPath:     urlPath,
			ResponseKey: responseKey,
			FieldsMap:   records,
			DocsURL:     docsURL,
		}

		return
	}

	if presentFields, isV1 := (any(object.FieldsMap)).(ObjectFieldsV1); isV1 {
		fieldsMap := make(ObjectFieldsV1)
		for k, v := range presentFields {
			fieldsMap[k] = v
		}

		for k, v := range (any(records)).(ObjectFieldsV1) { // nolint:forcetypeassert
			fieldsMap[k] = v
		}

		object.FieldsMap = any(fieldsMap).(F) // nolint:forcetypeassert

		return
	}

	if presentFields, isV2 := (any(object.FieldsMap)).(ObjectFieldsV2); isV2 {
		fieldsMap := make(ObjectFieldsV2)
		for k, v := range presentFields {
			fieldsMap[k] = v
		}

		for k, v := range (any(records)).(ObjectFieldsV2) { // nolint:forcetypeassert
			fieldsMap[k] = v
		}

		object.FieldsMap = any(fieldsMap).(F) // nolint:forcetypeassert

		return
	}
}

func (r *Metadata[F]) refactorLongestCommonPath() {
	for moduleID, module := range r.Modules {
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
		r.Modules[moduleID] = module
	}
}

func (r *Metadata[F]) getOrCreateModule(moduleID common.ModuleID) Module[F] {
	module, ok := r.Modules[moduleID]
	if !ok {
		// new module
		module = Module[F]{
			ID:      moduleID,
			Path:    "",
			Objects: make(map[string]Object[F]),
		}
		r.Modules[moduleID] = module
	}

	return module
}

// ObjectNames provides a registry of object names grouped by module.
func (r *Metadata[F]) ObjectNames() datautils.UniqueLists[common.ModuleID, string] {
	moduleObjectNames := make(datautils.UniqueLists[common.ModuleID, string])

	for key, value := range r.Modules {
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
func (r *Metadata[F]) LookupURLPath(moduleID common.ModuleID, objectName string) (string, error) {
	moduleID = moduleIdentifier(moduleID)

	path := r.Modules[moduleID].Objects[objectName].URLPath
	if len(path) == 0 {
		return "", common.ErrResolvingURLPathForObject
	}

	fullPath := r.LookupModuleURLPath(moduleID) + path

	return fullPath, nil
}

func (r *Metadata[F]) LookupModuleURLPath(moduleID common.ModuleID) string {
	moduleID = moduleIdentifier(moduleID)

	return r.Modules[moduleID].Path
}

// ModuleRegistry returns the list of API modules from static schema.
func (r *Metadata[F]) ModuleRegistry() common.Modules {
	result := make(common.Modules, len(r.Modules))

	for id, module := range r.Modules {
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
func (r *Metadata[F]) LookupArrayFieldName(moduleID common.ModuleID, objectName string) string {
	moduleID = moduleIdentifier(moduleID)

	fieldName := r.Modules[moduleID].Objects[objectName].ResponseKey

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
