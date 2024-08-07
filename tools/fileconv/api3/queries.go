package api3

import (
	"sort"
	"strings"
)

// Explorer allows to traverse schema in most common ways
// relevant for connectors metadata extraction.
type Explorer struct {
	schema *Document
}

func (e Explorer) GetBasicReadObjects(urlPrefix string, aliases Aliases, check ObjectCheck) ([]Schema, error) {
	schemas := make([]Schema, 0)

	for _, path := range e.GetBasicPathItems(urlPrefix) {
		schema, found, err := path.RetrieveSchemaOperationGet(aliases, check)
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
func (e Explorer) GetBasicPathItems(urlPrefix string) []PathItem {
	urlPrefix = formatPrefix(urlPrefix)

	result := make([]PathItem, 0)

	for path, pathObj := range e.schema.GetPaths() {
		if objectName, ok := strings.CutPrefix(path, urlPrefix); ok {
			if !strings.Contains(objectName, "/") &&
				!strings.Contains(objectName, "{") {
				// as of now only single word objects are supported
				// there should be no slashes, curly brackets - nested resources
				result = append(result, PathItem{
					name:     objectName,
					fullName: path,
					delegate: pathObj,
				})
			}
		}
	}

	return result
}

func formatPrefix(prefix string) string {
	// Prefix must start and end with slash.
	// Whole prefix as a single slash passes this requirement.
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return prefix
}
