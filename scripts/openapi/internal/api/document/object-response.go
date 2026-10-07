package document

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/getkin/kin-openapi/openapi3"
)

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
func extractItemsFromArrayHolder(
	endpointPath string, schema *openapi3.Schema, locator ArrayLocator,
	autoSelectArrayItem bool,
) (object *listObject, err error) {
	arrayOptions := extractPropertiesArrayType(schema)

	// Only one array property exists. We can conclude that this is the array item we are looking for.
	// Otherwise, match object name with target field name.
	approved := autoSelectArrayItem && len(arrayOptions) == 1

	for _, option := range arrayOptions {
		// Verify with the discriminator whether this is the target "Array".
		if approved || locator(endpointPath, option.Name) {
			return &listObject{
				endpointPath:   endpointPath,
				responseSchema: schema,
				itemsSchema:    option.Item.Value,
				itemsSchemaRef: option.Item.Ref,
				itemsLocation:  option.Name,
			}, nil
		}
	}

	if isBooleanTruthful(schema.AdditionalProperties.Has) {
		// This schema is dynamic.
		// the fields cannot be known.
		return &listObject{
			endpointPath:   endpointPath,
			responseSchema: schema,
			isDynamicItems: true,
		}, nil
	}

	return nil, createUnprocessableObjectError(endpointPath)
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
