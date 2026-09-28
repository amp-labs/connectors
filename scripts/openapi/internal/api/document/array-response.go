package document

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// Response schema is an array itself. Get schema describing items.
func extractItemsFromArray(endpointPath string,
	schema *openapi3.Schema,
) (*listObject, error) {
	items, isArray := getItems(schema.NewRef())
	if !isArray {
		return nil, createUnprocessableObjectError(endpointPath)
	}

	return &listObject{
		endpointPath:   endpointPath,
		responseSchema: schema,
		itemsSchema:    items.Value,
		itemsSchemaRef: items.Ref,
		itemsLocation:  "",
	}, nil
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
