package api3

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
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
	objectName string
	urlPath    string
	delegate   *openapi3.PathItem
}

type Schema struct {
	ObjectName  string
	DisplayName string
	Fields      []string
	QueryParams []string
	URLPath     string
	Problem     error
}

type Schemas []Schema

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

	return fmt.Sprintf("%v=[%v]", s.ObjectName, strings.Join(s.Fields, ","))
}

func (p PathItem) RetrieveSchemaOperation(
	operationName string,
	displayNameOverride map[string]string, check ObjectCheck, displayProcessor DisplayNameProcessor,
	parameterFilter ParameterFilterGetMethod,
) (*Schema, bool, error) {
	operation := p.selectOperation(operationName)
	if operation == nil {
		return nil, false, nil
	}

	if parameterFilter != nil {
		ok := parameterFilter(p.objectName, operation)
		if !ok {
			// Omit this schema. We only work with GET method without required parameters
			return nil, false, nil
		}
	}

	schema := extractSchema(operation)
	if schema == nil {
		return nil, false, nil
	}

	displayName, ok := displayNameOverride[p.objectName]
	if !ok {
		displayName = schema.Title
		if len(displayName) == 0 {
			displayName = p.objectName
		}

		if displayProcessor != nil {
			// Post process Display Names to have shared format.
			displayName = displayProcessor(displayName)
		}
	}

	fields, err := extractFieldsFromArrayItem(p.objectName, schema, check)

	return &Schema{
		ObjectName:  p.objectName,
		DisplayName: displayName,
		Fields:      fields,
		QueryParams: getQueryParameters(operation),
		URLPath:     p.urlPath,
		Problem:     err,
	}, true, nil
}

func (p PathItem) selectOperation(operationName string) *openapi3.Operation {
	switch operationName {
	case "POST":
		return p.delegate.Post
	case "PUT":
		return p.delegate.Put
	case "PATCH":
		return p.delegate.Patch
	default:
		return p.delegate.Get
	}
}

func extractFieldsFromArrayItem(objectName string, schema *openapi3.Schema, check ObjectCheck) ([]string, error) {
	checkExpectationsOperationResponseSchema(schema)

	definitions := []openapi3.Schemas{
		schema.Properties,
	}
	for _, allOf := range schema.AllOf {
		// item schema will likely be in composite schema
		definitions = append(definitions, allOf.Value.Properties)
	}

	for _, definition := range definitions {
		for name, nestedSchema := range definition {
			if items, ok := getItems(nestedSchema); ok {
				// We are interested in the schema of array type.
				// Those fields of an item are what we are after.
				// Now ask the discriminator if this is the target List.
				// It is possible that response has multiple arrays, that's why we are asking to resolve ambiguity.
				if check(objectName, name) {
					return extractFields(items.Value)
				}
			}
		}
	}

	if goutils.Pointers.IsTrue(schema.AdditionalProperties.Has) {
		// this schema is dynamic.
		// the fields cannot be known.
		return []string{}, nil
	}

	return nil, fmt.Errorf("%w: object %v", ErrUnprocessableObject, objectName)
}

func getItems(schema *openapi3.SchemaRef) (*openapi3.SchemaRef, bool) {
	if schema.Value == nil {
		return nil, false
	}

	if schema.Value.Items == nil {
		return nil, false
	}

	return schema.Value.Items, true
}

func getQueryParameters(operation *openapi3.Operation) []string {
	queryParams := make([]string, 0, len(operation.Parameters))

	for _, parameter := range operation.Parameters {
		if parameter.Value.In == "query" {
			queryParams = append(queryParams, parameter.Value.Name)
		}
	}

	return queryParams
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
	combined := make(datautils.Set[string])

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

// This logs any concerns if any.
// The OpenAPI extractor has some expectations that should hold true, otherwise the extraction
// should be rethought to match the edge case.
//
// Operation response has a schema, it must be an object, and it should contain a field
// that will hold Object of interest. That object is what connectors.ReadConnector returns via Read method.
func checkExpectationsOperationResponseSchema(schema *openapi3.Schema) {
	if schema.Type == nil {
		slog.Warn("Schema definition has no type")

		return
	}

	if len(*schema.Type) != 1 {
		slog.Warn("Schema definition has multiple types")
	}

	found := false

	for _, s := range *schema.Type {
		if s == "object" {
			found = true
		}
	}

	if !found {
		slog.Warn("Schema definition is not an object. Expected to be an object containing array of items.")
	}
}
