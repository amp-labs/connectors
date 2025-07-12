package metadatadef

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

// Schema is a model that describes a REST API object.
// This is usually created when metadata is coming not from API but alternative sources, example: OpenAPI.
// This model holds more information than common.ObjectMetadata.
// It may include custom properties.
type Schema = ExtendedSchema[any]

type ExtendedSchema[C any] struct {
	ObjectName  string
	DisplayName string
	Fields      Fields
	QueryParams []string
	URLPath     string
	ResponseKey string
	Problem     error
	Custom      C
}

type Schemas[C any] []ExtendedSchema[C]

type Field struct {
	Name         string
	DisplayName  string
	Type         string
	ValueOptions []string
}

type Fields = datautils.Map[string, Field]

func (s Schemas[C]) Combine(others Schemas[C]) Schemas[C] {
	registry := datautils.Map[string, ExtendedSchema[C]]{}
	for _, schema := range append(s, others...) {
		_, found := registry[schema.ObjectName]

		if !found || len(schema.Fields) != 0 {
			registry[schema.ObjectName] = schema
		}
	}

	return registry.Values()
}

func (s ExtendedSchema[C]) String() string {
	if s.Problem != nil {
		return fmt.Sprintf("    {%v}    ", s.ObjectName)
	}

	return fmt.Sprintf("%v=[%v]", s.ObjectName, strings.Join(s.Fields.Keys(), ","))
}
