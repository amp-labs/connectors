package staticschema

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
)

const RootModuleID common.ModuleID = "root"

// NewMetadata constructs empty Metadata. To populate it with data use Metadata.Add method.
func NewMetadata() *Metadata {
	return &Metadata{
		Modules: make(map[common.ModuleID]Module),
	}
}

// Metadata is data that is saved in the static file.
//
// This structure offers the following main methods:
// Metadata.ObjectNames - registry of object names by modules.
// Metadata.LookupURLPath - resolves URL given object name.
// Metadata.Select - produces output suitable for connectors.ObjectMetadataConnector.
type Metadata struct {
	// Modules is a map of object names to object metadata
	Modules map[common.ModuleID]Module `json:"modules"`
}

type Module struct {
	ID      common.ModuleID   `json:"id"`
	Path    string            `json:"path"`
	Objects map[string]Object `json:"objects"`
}

type Object struct {
	// Provider's display name for the object
	DisplayName string `json:"displayName"`

	// This is the endpoint URLPath to this REST API resource/object.
	// It can be used to make request to "list objects".
	URLPath string `json:"path"`

	// FieldsMap is a map of field names to field display names
	FieldsMap map[string]string `json:"fields"`

	// DocsURL points to docs endpoint. Optional.
	DocsURL *string `json:"docs,omitempty"`
}

// Add will appropriately store the data, abiding to data structure rules.
// NOTE: empty module id is treated as root module.
func (r *Metadata) Add(
	moduleID common.ModuleID,
	objectName, objectDisplayName, fieldName, urlPath string,
	docsURL *string,
) {
	moduleID = moduleIdentifier(moduleID)

	module := r.getOrCreateModule(moduleID)

	data, ok := module.Objects[objectName]
	if !ok {
		data = Object{
			DisplayName: objectDisplayName,
			URLPath:     urlPath,
			FieldsMap:   make(map[string]string),
			DocsURL:     docsURL,
		}
		module.Objects[objectName] = data
	}

	data.FieldsMap[fieldName] = fieldName
}

func (r *Metadata) refactorLongestCommonPath() {
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

func (r *Metadata) getOrCreateModule(moduleID common.ModuleID) Module {
	module, ok := r.Modules[moduleID]
	if !ok {
		// new module
		module = Module{
			ID:      moduleID,
			Path:    "",
			Objects: make(map[string]Object),
		}
		r.Modules[moduleID] = module
	}

	return module
}

// ObjectNames provides a registry of object names grouped by module.
func (r *Metadata) ObjectNames() handy.UniqueLists[common.ModuleID, string] {
	moduleObjectNames := make(handy.UniqueLists[common.ModuleID, string])

	for key, value := range r.Modules {
		names := handy.NewStringSet()
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
func (r *Metadata) LookupURLPath(moduleID common.ModuleID, objectName string) (string, error) {
	moduleID = moduleIdentifier(moduleID)

	path := r.Modules[moduleID].Objects[objectName].URLPath
	if len(path) == 0 {
		return "", common.ErrResolvingURLPathForObject
	}

	prefix := r.Modules[moduleID].Path
	fullPath := prefix + path

	return fullPath, nil
}

// ModuleRegistry returns the list of API modules from static schema.
func (r *Metadata) ModuleRegistry() common.Modules {
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

func (m *Module) withPath(path string) {
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

	for i := 0; i < shortestLength; i++ {
		if first[i] != second[i] {
			return result
		}

		result += string(first[i])
	}

	return result
}
