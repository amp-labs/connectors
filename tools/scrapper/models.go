package scrapper

import (
	"sort"
	"strings"
)

type ModelDocLinks []ModelDocLink

func (l ModelDocLinks) FindByName(name string) (ModelDocLink, bool) {
	for _, link := range l {
		if link.Name == name {
			return link, true
		}
	}

	return ModelDocLink{}, false
}

type ModelURLRegistry struct {
	ModelDocs ModelDocLinks `json:"data"`
}

type ModelDocLink struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
}

func NewModelURLRegistry() *ModelURLRegistry {
	return &ModelURLRegistry{
		ModelDocs: make([]ModelDocLink, 0),
	}
}

func (r *ModelURLRegistry) Add(displayName, url string) {
	if len(displayName) == 0 || len(url) == 0 {
		// Trying to add URL with no display name or missing URL
		return
	}

	url, _ = strings.CutSuffix(url, "/")
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]

	r.ModelDocs = append(r.ModelDocs, ModelDocLink{
		DisplayName: displayName,
		Name:        name,
		URL:         url,
	})
}

func (r *ModelURLRegistry) Sort() {
	sort.Slice(r.ModelDocs, func(i, j int) bool {
		return r.ModelDocs[i].Name < r.ModelDocs[j].Name
	})
}

type ObjectMetadataResult struct {
	// Result is a map of object names to object metadata
	Result map[string]ObjectMetadata `json:"data"`
}

type ObjectMetadata struct {
	// Provider's display name for the object
	DisplayName string `json:"displayName"`

	// FieldsMap is a map of field names to field display names
	FieldsMap map[string]string `json:"fields"`
}

func NewObjectMetadataResult() *ObjectMetadataResult {
	return &ObjectMetadataResult{
		Result: make(map[string]ObjectMetadata),
	}
}

func (r *ObjectMetadataResult) Add(objectName string, objectDisplayName string, fieldName string) {
	data, ok := r.Result[objectName]
	if !ok {
		data = ObjectMetadata{
			DisplayName: objectDisplayName,
			FieldsMap:   make(map[string]string),
		}
		r.Result[objectName] = data
	}

	data.FieldsMap[fieldName] = fieldName
}
