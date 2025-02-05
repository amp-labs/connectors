package metadatadef

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

// Schema is a model that describes a REST API object.
// This is usually created when metadata is coming not from API but alternative sources, example: OpenAPI.
// This model holds more information than common.ObjectMetadata.
type Schema struct {
	ObjectName  string
	DisplayName string
	Fields      Fields
	QueryParams []string
	URLPath     string
	ResponseKey string
	Problem     error
}

type Schemas []Schema

type Field struct {
	Name        string
	Type        string
	EnumOptions []string
}

type Fields = datautils.Map[string, Field]

func (s Schemas) Combine(others Schemas) Schemas {
	registry := datautils.Map[string, Schema]{}
	for _, schema := range append(s, others...) {
		_, found := registry[schema.ObjectName]

		if !found || len(schema.Fields) != 0 {
			registry[schema.ObjectName] = schema
		}
	}

	return registry.Values()
}

func (s Schema) String() string {
	if s.Problem != nil {
		return fmt.Sprintf("    {%v}    ", s.ObjectName)
	}

	return fmt.Sprintf("%v=[%v]", s.ObjectName, strings.Join(s.Fields.Keys(), ","))
}
