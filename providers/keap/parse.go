package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/keap/metadata"
	"github.com/spyzhov/ajson"
)

func makeGetRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)

		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	page, err := jsonquery.New(node).StrWithDefault("next", "")
	if err != nil {
		return "", err
	}

	if page != "" {
		// API Version 1.
		return page, nil
	}

	// API Version 2, try to look for next page in a different place.
	return jsonquery.New(node).StrWithDefault("next_page_token", "")
}

// Before parsing the records, if any custom fields are present (without a human-readable name),
// this will call the correct API to extend & replace the custom field with human-readable information.
// Object will then be enhanced using model.
func (c *Connector) attachReadCustomFields(customFields map[string]modelCustomField) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		if len(customFields) == 0 {
			// No custom fields, no-op, return as is.
			return jsonquery.Convertor.ObjectToMap(node)
		}

		return enhanceObjectsWithCustomFieldNames(node, customFields)
	}
}

// In general this does the usual JSON parsing.
// However, those objects that contain "custom_fields" are processed as follows:
// * Locate custom fields in JSON read response.
// * Replace ids with human-readable names, which is provided as argument.
// * Place fields at the top level of the object.
func enhanceObjectsWithCustomFieldNames(
	node *ajson.Node,
	fields map[string]modelCustomField,
) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsResponse, err := jsonquery.ParseNode[readCustomFieldsResponse](node)
	if err != nil {
		return nil, err
	}

	// Replace identifiers with human-readable field names which were found by making a call to "/model".
	for _, field := range customFieldsResponse.CustomFields {
		if model, ok := fields[field.ID.String()]; ok {
			object[model.Name()] = field.Content
		}
	}

	return object, nil
}
