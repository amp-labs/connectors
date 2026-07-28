package document

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/getkin/kin-openapi/openapi3"
)

type HTTPOperation string

var ErrUnprocessableObject = errors.New("don't know how to process schema")

// Document is a wrapper of openapi with null checks.
type Document struct {
	delegate *openapi3.T
}

func New(data *openapi3.T) *Document {
	return &Document{
		delegate: data,
	}
}

func (s Document) GetPaths() []PathItem {
	paths := s.delegate.Paths
	if paths == nil {
		return nil
	}

	result := make([]PathItem, 0)
	for urlPath, pathItem := range paths.Map() {
		result = append(result, PathItem{
			urlPath:  urlPath,
			delegate: pathItem,
		})
	}

	return result
}

type PathItem struct {
	urlPath  string
	delegate *openapi3.PathItem
}

func (p PathItem) RetrieveSchemaOperation(
	httpOperation HTTPOperation,
	locator ArrayLocator,
	propertyFlattener PropertyFlattener,
	mime string,
	autoSelectArrayItem bool,
) (*spec.Schema, bool, error) {
	operation, _ := p.selectOperation(httpOperation)
	if operation == nil {
		return nil, false, nil
	}

	schema := extractSchema(operation, mime)
	if schema == nil {
		return nil, false, nil
	}

	fields, responseKey, err := extractObjectFields(p.urlPath, schema, locator, propertyFlattener, autoSelectArrayItem)
	if err == nil && len(fields) == 0 {
		slog.Info("not an array of objects", "endpoint", p.urlPath)
	}

	objectName, _ := strings.CutPrefix(p.urlPath, "/")

	return &spec.Schema{
		ObjectName:  objectName,
		URLPath:     p.urlPath,
		Operation:   string(httpOperation),
		DisplayName: schema.Title,
		ResponseKey: responseKey,
		Fields:      fields,
		QueryParams: getQueryParameters(operation),
		Problem:     err,
		Raw:         schema,
	}, true, nil
}

func (p PathItem) selectOperation(httpOperation HTTPOperation) (*openapi3.Operation, bool) {
	switch httpOperation {
	case http.MethodGet:
		return p.delegate.Get, p.delegate.Get != nil
	case http.MethodPost:
		return p.delegate.Post, p.delegate.Post != nil
	case http.MethodPut:
		return p.delegate.Put, p.delegate.Put != nil
	case http.MethodPatch:
		return p.delegate.Patch, p.delegate.Patch != nil
	case http.MethodDelete:
		return p.delegate.Delete, p.delegate.Delete != nil
	default:
		// Bool will always be false for the default case. This operation is not what we are looking for.
		return p.delegate.Get, false
	}
}

func extractObjectFields(
	endpointPath string, schema *openapi3.Schema, locator ArrayLocator,
	propertyFlattener PropertyFlattener,
	autoSelectArrayItem bool,
) (fields spec.Fields, location string, err error) {
	switch getSchemaType(endpointPath, schema) {
	case schemaTypeObject:
		return extractFieldsFromArrayHolder(endpointPath, schema, locator, propertyFlattener, autoSelectArrayItem)
	case schemaTypeArray:
		return extractFieldsFromArray(endpointPath, schema, propertyFlattener)
	case schemaTypeUnknown:
		// Even though OpenAPI doesn't explicitly state that the schema is of object type,
		// Attempt to process it as such.
		// It seems that some OpenAPI files are not that strict about such things. Ex: Pipedrive, Zendesk.
		return extractFieldsFromArrayHolder(endpointPath, schema, locator, propertyFlattener, autoSelectArrayItem)
	default:
		return nil, "", createUnprocessableObjectError(endpointPath)
	}
}

type Array struct {
	Name string
	Item *openapi3.SchemaRef
}

// The response schema is an object that contains one or more arrays of items.
// It is not immediately clear which field holds the list of items we are interested in.
// Since the object may have multiple array properties, we cannot assume a single array
// will always be present. To resolve this ambiguity, the ArrayLocator callback
// is used to select the appropriate array.
//
// Example:
//
//	{
//	    "products": [...],
//	    "prices": [...],
//	    "links": {
//	        "next": "url"
//	    }
//	}
//
// In this case, it is unclear whether to use the "products" or "prices" schema.
// The ArrayLocator will determine the correct field.
//
// Alternatively, if only one array is present and autoSelectArrayItem is enabled,
// the selection will happen automatically without invoking the callback.
func extractFieldsFromArrayHolder(
	endpointPath string, schema *openapi3.Schema, locator ArrayLocator,
	propertyFlattener PropertyFlattener,
	autoSelectArrayItem bool,
) (fields spec.Fields, location string, err error) {
	arrayOptions := extractPropertiesArrayType(schema)

	// Only one array property exists. We can conclude that this is the array item we are looking for.
	// Otherwise, match object name with target field name.
	approved := autoSelectArrayItem && len(arrayOptions) == 1

	for _, option := range arrayOptions {
		// Verify with the discriminator whether this is the target "Array".
		if approved || locator(endpointPath, option.Name) {
			fields, err = extractFields(endpointPath, propertyFlattener, option.Item.Value)
			if err != nil {
				return nil, "", err
			}

			return fields, option.Name, nil
		}
	}

	if isBooleanTruthful(schema.AdditionalProperties.Has) {
		// this schema is dynamic.
		// the fields cannot be known.
		return make(spec.Fields), "", nil
	}

	return nil, "", createUnprocessableObjectError(endpointPath)
}

// The schema contains multiple properties that collectively form a normalized representation of the API response.
// The procedure will identify and collect only the fields of array type.
func extractPropertiesArrayType(schema *openapi3.Schema) []Array {
	definitions := []openapi3.Schemas{
		schema.Properties,
	}

	compositeSchemas := datautils.MergeSlices(schema.AllOf, schema.OneOf, schema.AnyOf)
	// Item schema will likely be inside composite schema
	for _, comp := range compositeSchemas {
		if comp != nil && comp.Value != nil {
			definitions = append(definitions, comp.Value.Properties)
		}
	}

	arrays := make([]Array, 0)
	// Collect properties that are of an array type.
	for _, definition := range definitions {
		for name, nestedSchema := range definition {
			if itemsSchema, isArray := getItems(nestedSchema); isArray {
				arrays = append(arrays, Array{
					Name: name,
					Item: itemsSchema,
				})
			}
		}
	}

	return arrays
}

// Response schema is an array itself. Collect fields that describe single item.
func extractFieldsFromArray(
	endpointPath string, schema *openapi3.Schema,
	propertyFlattener PropertyFlattener,
) (fields spec.Fields, location string, err error) {
	items, isArray := getItems(schema.NewRef())
	if !isArray {
		return nil, "", createUnprocessableObjectError(endpointPath)
	}

	fields, err = extractFields(endpointPath, propertyFlattener, items.Value)

	return fields, "", err
}

func getItems(schema *openapi3.SchemaRef) (itemsSchema *openapi3.SchemaRef, isArray bool) {
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
	endpointPath string,
	propertyFlattener PropertyFlattener, source *openapi3.Schema,
) (spec.Fields, error) {
	combinedFields := make(spec.Fields)

	if source.AnyOf != nil {
		// this object can be represented by various definitions
		// we merge those fields to represent the whole domain of possible fields
		// of course omitting duplicates.
		for _, ref := range source.AnyOf {
			fields, err := extractFields(endpointPath, propertyFlattener, ref.Value)
			if err != nil {
				return nil, err
			}

			combinedFields.AddMapValues(fields)
		}
	}

	// for all parents that exist collect fields
	for _, ref := range source.AllOf {
		parentValue := ref.Value
		if parentValue != nil {
			fields, err := extractFields(endpointPath, propertyFlattener, parentValue)
			if err != nil {
				return nil, err
			}

			combinedFields.AddMapValues(fields)
		}
	}

	// properties local to this schema
	for property, propertySchema := range source.Properties {
		if propertyFlattener(endpointPath, property) {
			// This property holds an array, and we need nested fields to be moved one level up.
			fields, err := extractFields(endpointPath, propertyFlattener, propertySchema.Value)
			if err != nil {
				return nil, err
			}

			combinedFields.AddMapValues(fields)
		} else {
			// This is just a normal usual case where top level fields are collected as is.
			propertyType := extractPropertyType(propertySchema)
			enumOptions := extractEnumOptions(endpointPath, propertySchema)
			combinedFields[property] = spec.Field{
				Name:         property,
				Type:         propertyType,
				ValueOptions: enumOptions,
			}
		}
	}

	return combinedFields, nil
}

func extractPropertyType(propertySchema *openapi3.SchemaRef) string {
	if propertySchema.Value == nil || propertySchema.Value.Type == nil {
		return ""
	}

	types := *propertySchema.Value.Type
	if len(types) == 0 {
		return ""
	}

	return types[0]
}

func extractEnumOptions(endpointPath string, propertySchema *openapi3.SchemaRef) []string {
	enumOptions := make([]string, 0)

	if propertySchema.Value != nil && propertySchema.Value.Enum != nil {
		for _, value := range propertySchema.Value.Enum {
			if option, ok := mapEnumToString(endpointPath, value); ok {
				enumOptions = append(enumOptions, option)
			}
		}
	}

	return enumOptions
}

func mapEnumToString(endpointPath string, value any) (string, bool) {
	if value == nil {
		return "null", true
	}

	switch option := value.(type) {
	case string:
		return option, true
	case float64:
		return strconv.FormatFloat(option, 'f', -1, 64), true
	default:
		slog.Warn("Enum option is not a string", "endpoint", endpointPath)
	}

	return "", false
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
func getSchemaType(endpointPath string, schema *openapi3.Schema) definitionSchemaType {
	if schema.Type == nil {
		slog.Info("Schema definition has no type", "endpoint", endpointPath)

		return schemaTypeUnknown
	}

	if len(*schema.Type) != 1 {
		slog.Info("Schema definition has multiple types")
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

	slog.Info("Schema definition is neither an object nor an array. " +
		"Expected to be an object containing array of items, or the list itself.")

	return schemaTypeUnknown
}

func createUnprocessableObjectError(endpointPath string) error {
	return fmt.Errorf("%w: object path %v", ErrUnprocessableObject, endpointPath)
}

func isBooleanTruthful(input *bool) bool {
	if input == nil {
		return false
	}

	return *input
}
