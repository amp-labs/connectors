package api3

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common/handy"
	"github.com/getkin/kin-openapi/openapi3"
)

var ErrUnprocessableObject = errors.New("don't know how to process schema")

// Document is a wrapper of openapi with null checks.
type Document struct {
	delegate *openapi3.T
}

func (s Document) GetPaths() map[string]*openapi3.PathItem {
	paths := s.delegate.Paths
	if paths == nil {
		return nil
	}

	return paths.Map()
}

type PathItem struct {
	name     string
	fullName string
	delegate *openapi3.PathItem
}

type Schema struct {
	ObjectName  string
	DisplayName string
	Fields      []string
	Problem     error
}

func (s Schema) String() string {
	if s.Problem != nil {
		return fmt.Sprintf("    {%v}    ", s.ObjectName)
	}

	return fmt.Sprintf("%v=[%v]", s.ObjectName, strings.Join(s.Fields, ","))
}

func (p PathItem) RetrieveSchemaOperationGet(aliases Aliases, check ObjectCheck) (*Schema, bool, error) {
	operation := p.delegate.Get
	if operation == nil {
		return nil, false, nil
	}

	schema := extractSchema(operation)
	if schema == nil {
		return nil, false, nil
	}

	name := aliases.Synonym(p.name)
	displayName := schema.Title

	if len(displayName) == 0 {
		displayName = name
	}

	fields, err := extractFieldsFromArrayItem(name, schema, check)

	return &Schema{
		ObjectName:  name,
		DisplayName: displayName,
		Fields:      fields,
		Problem:     err,
	}, true, nil
}

func extractFieldsFromArrayItem(objectName string, schema *openapi3.Schema, check ObjectCheck) ([]string, error) {
	definitions := []openapi3.Schemas{
		schema.Properties,
	}
	for _, allOf := range schema.AllOf {
		// item schema will likely be in composite schema
		definitions = append(definitions, allOf.Value.Properties)
	}

	for _, definition := range definitions {
		for name, nestedSchema := range definition {
			if check(name, objectName) {
				// Object was found.
				// We interested in the schema of array type.
				// Those fields of an item are what we are after.
				items := nestedSchema.Value.Items
				if items != nil {
					return extractFields(items.Value)
				}
			}
		}
	}

	if handy.Pointers.IsTrue(schema.AdditionalProperties.Has) {
		// this schema is dynamic.
		// the fields cannot be known.
		return []string{}, nil
	}

	return nil, fmt.Errorf("%w: object %v", ErrUnprocessableObject, objectName)
}

func extractSchema(operation *openapi3.Operation) *openapi3.Schema {
	if operation == nil {
		return nil
	}

	responses := operation.Responses
	if responses == nil {
		return nil
	}

	// any 2xx response will suffice
	status := responses.Status(http.StatusOK)
	if status == nil {
		return nil
	}

	value := status.Value
	if value == nil {
		return nil
	}

	mediaType := value.Content.Get("application/json")
	if mediaType == nil {
		return nil
	}

	schema := mediaType.Schema
	if schema == nil {
		return nil
	}

	schemaValue := schema.Value
	if schemaValue == nil {
		return nil
	}

	return schemaValue
}

func extractFields(source *openapi3.Schema) ([]string, error) {
	combined := make(handy.Set[string])

	if source.AnyOf != nil {
		// this object can be represented by various definitions
		// we merge those fields to represent the whole domain of possible fields
		// of course omitting duplicates.
		for _, ref := range source.AnyOf {
			fields, err := extractFields(ref.Value)
			if err != nil {
				return nil, err
			}

			combined.Add(fields)
		}
	}

	// for all parents that exist collect fields
	for _, ref := range source.AllOf {
		parentValue := ref.Value
		if parentValue != nil {
			fields, err := extractFields(parentValue)
			if err != nil {
				return nil, err
			}

			combined.Add(fields)
		}
	}

	// properties local to this schema
	for property := range source.Properties {
		combined.AddOne(property)
	}

	return combined.List(), nil
}
