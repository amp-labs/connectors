package api3

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
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
	ResponseKey string
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
	displayNameOverride map[string]string,
	locator ObjectArrayLocator,
	displayProcessor DisplayNameProcessor,
	operationMethodFilter ReadOperationMethodFilter,
	propertyFlattener PropertyFlattener,
	mime string,
) (*Schema, bool, error) {
	operation := p.selectOperation(operationName)
	if operation == nil {
		return nil, false, nil
	}

	if ok := operationMethodFilter(p.objectName, operation); !ok {
		// Omit this schema, operation for this object is not what we are looking for.
		return nil, false, nil
	}

	schema := extractSchema(operation, mime)
	if schema == nil {
		return nil, false, nil
	}

	displayName, ok := displayNameOverride[p.objectName]
	if !ok {
		displayName = schema.Title
		if len(displayName) == 0 {
			displayName = p.objectName
		}

		// Post process Display Names to have shared format.
		displayName = displayProcessor(displayName)
	}

	fields, responseKey, err := extractObjectFields(p.objectName, schema, locator, propertyFlattener)

	return &Schema{
		ObjectName:  p.objectName,
		DisplayName: displayName,
		Fields:      fields,
		QueryParams: getQueryParameters(operation),
		URLPath:     p.urlPath,
		ResponseKey: responseKey,
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

func extractObjectFields(
	objectName string, schema *openapi3.Schema, locator ObjectArrayLocator,
	propertyFlattener PropertyFlattener,
) (fields []string, location string, err error) {
	switch getSchemaType(schema) {
	case schemaTypeObject:
		return extractFieldsFromArrayHolder(objectName, schema, locator, propertyFlattener)
	case schemaTypeArray:
		return extractFieldsFromArray(objectName, schema, propertyFlattener)
	case schemaTypeUnknown:
		// Even though OpenAPI doesn't explicitly state that the schema is of object type,
		// Attempt to process it as such.
		// It seems that some OpenAPI files are not that strict about such things. Ex: Pipedrive, Zendesk.
		return extractFieldsFromArrayHolder(objectName, schema, locator, propertyFlattener)
	default:
		return nil, "", createUnprocessableObjectError(objectName)
	}
}

// Response schema is an object. This object holds an array of items.
// It is not obvious which field holds the list of items that we are interested in.
// We cannot assume that object will always have one property of array type, therefore an ObjectArrayLocator is used
// to determine what is the name of property with respect to the object name.
func extractFieldsFromArrayHolder(
	objectName string, schema *openapi3.Schema, locator ObjectArrayLocator,
	propertyFlattener PropertyFlattener,
) (fields []string, location string, err error) {
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
				if locator(objectName, name) {
					fields, err = extractFields(objectName, propertyFlattener, items.Value)
					if err != nil {
						return nil, "", err
					}

					return fields, name, nil
				}
			}
		}
	}

	if isBooleanTruthful(schema.AdditionalProperties.Has) {
		// this schema is dynamic.
		// the fields cannot be known.
		return []string{}, "", nil
	}

	return nil, "", createUnprocessableObjectError(objectName)
}

// Response schema is an array itself. Collect fields that describe single item.
func extractFieldsFromArray(
	objectName string, schema *openapi3.Schema,
	propertyFlattener PropertyFlattener,
) (fields []string, location string, err error) {
	items, ok := getItems(schema.NewRef())
	if !ok {
		return nil, "", createUnprocessableObjectError(objectName)
	}

	fields, err = extractFields(objectName, propertyFlattener, items.Value)

	return fields, "", err
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

func extractSchema(operation *openapi3.Operation, mime string) *openapi3.Schema {
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

	mediaType := value.Content.Get(mime)
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

func extractFields(
	objectName string,
	propertyFlattener PropertyFlattener, source *openapi3.Schema,
) ([]string, error) {
	combined := make(datautils.Set[string])

	if source.AnyOf != nil {
		// this object can be represented by various definitions
		// we merge those fields to represent the whole domain of possible fields
		// of course omitting duplicates.
		for _, ref := range source.AnyOf {
			fields, err := extractFields(objectName, propertyFlattener, ref.Value)
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
			fields, err := extractFields(objectName, propertyFlattener, parentValue)
			if err != nil {
				return nil, err
			}

			combined.Add(fields)
		}
	}

	// properties local to this schema
	for property, propertySchema := range source.Properties {
		if propertyFlattener(objectName, property) {
			// This property holds an array, and we need nested fields to be moved one level up.
			fields, err := extractFields(objectName, propertyFlattener, propertySchema.Value)
			if err != nil {
				return nil, err
			}

			combined.Add(fields)
		} else {
			// This is just a normal usual case where top level fields are collected as is.
			combined.AddOne(property)
		}
	}

	return combined.List(), nil
}

type definitionSchemaType int

const (
	schemaTypeUnknown definitionSchemaType = iota
	schemaTypeObject
	schemaTypeArray
)

// This logs any concerns if any.
// The OpenAPI extractor has some expectations that should hold true, otherwise the extraction
// should be rethought to match the edge case.
//
// Operation response has a schema, it must be an object, and it should contain a field
// that will hold Object of interest. That object is what connectors.ReadConnector returns via Read method.
// It is allowed to have an array without it being nested under object.
//
// Returns enum marking the type of schema. This can be used to adjust processing.
func getSchemaType(schema *openapi3.Schema) definitionSchemaType {
	if schema.Type == nil {
		slog.Warn("Schema definition has no type")

		return schemaTypeUnknown
	}

	if len(*schema.Type) != 1 {
		slog.Warn("Schema definition has multiple types")
	}

	for _, s := range *schema.Type {
		if s == "object" {
			return schemaTypeObject
		}
	}

	for _, s := range *schema.Type {
		if s == "array" {
			return schemaTypeArray
		}
	}

	slog.Warn("Schema definition is neither an object nor an array. " +
		"Expected to be an object containing array of items, or the list itself.")

	return schemaTypeUnknown
}

func createUnprocessableObjectError(objectName string) error {
	return fmt.Errorf("%w: object %v", ErrUnprocessableObject, objectName)
}

func isBooleanTruthful(input *bool) bool {
	if input == nil {
		return false
	}

	return *input
}
