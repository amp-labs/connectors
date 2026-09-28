package document

import (
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/getkin/kin-openapi/openapi3"
)

func (p PathItem) RetrieveListResponseSchema(httpOperation HTTPOperation,
	locator ArrayLocator,
	propertyFlattener PropertyFlattener,
	operationFilter OperationFilter,
	mime string,
	autoSelectArrayItem bool,
) (*spec.Schema, bool, error) {
	operation, _ := p.selectOperation(httpOperation)
	if operation == nil {
		return nil, false, nil
	}

	if ok := operationFilter(p.urlPath, operation); !ok {
		// Omit this schema, operation for this object is not what we are looking for.
		return nil, false, nil
	}

	schema := extractSchema(operation, mime)
	if schema == nil {
		return nil, false, nil
	}

	objectName, _ := strings.CutPrefix(p.urlPath, "/")
	result := &spec.Schema{
		ObjectName:  objectName,
		UrlPath:     p.urlPath,
		Operation:   string(httpOperation),
		QueryParams: getQueryParameters(operation),
		DisplayName: schema.Title,
		Origin: spec.SchemaOrigin{
			ResponseSchema: schema,
			ItemsSchemaRef: "",
			ItemsSchema:    nil,
		},
	}

	object, err := extractListObjectForSchema(p.urlPath, schema, locator, autoSelectArrayItem)
	if err != nil {
		result.Problem = err

		return result, true, nil // nolint:nilerr
	}

	result.Fields = object.GetFields(propertyFlattener)
	result.ResponseKey = object.itemsLocation
	result.Origin.ItemsSchemaRef = object.itemsSchemaRef
	result.Origin.ItemsSchema = object.itemsSchema
	result.Origin.ItemComponentName = schemaRefComponentName(object.itemsSchemaRef)

	return result, true, nil
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

func extractListObjectForSchema(
	endpointPath string, schema *openapi3.Schema, locator ArrayLocator, autoSelectArrayItem bool,
) (object *listObject, err error) {
	switch getSchemaType(schema) {
	case schemaTypeObject:
		return extractItemsFromArrayHolder(endpointPath, schema, locator, autoSelectArrayItem)
	case schemaTypeArray:
		return extractItemsFromArray(endpointPath, schema)
	case schemaTypeUnknown:
		// Even though OpenAPI doesn't explicitly state that the schema is of object type,
		// Attempt to process it as such.
		// It seems that some OpenAPI files are not that strict about such things. Ex: Pipedrive, Zendesk.
		return extractItemsFromArrayHolder(endpointPath, schema, locator, autoSelectArrayItem)
	default:
		return nil, createUnprocessableObjectError(endpointPath)
	}
}

type listObject struct {
	endpointPath   string
	responseSchema *openapi3.Schema
	itemsSchema    *openapi3.Schema
	itemsSchemaRef string
	itemsLocation  string
	isDynamicItems bool
}

func (o *listObject) GetFields(propertyFlattener PropertyFlattener) spec.Fields {
	if o.isDynamicItems {
		return make(spec.Fields)
	}

	return extractFields(o.endpointPath, propertyFlattener, o.itemsSchema,
		fieldDefinitionLocation{
			schema: o.itemsSchema,
			paths:  []string{o.itemsSchemaRef},
		},
	)
}
