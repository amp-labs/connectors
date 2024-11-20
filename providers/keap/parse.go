package keap

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(moduleID common.ModuleID) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if moduleID == ModuleV1 {
			return jsonquery.New(node).StrWithDefault("next", "")
		}

		return jsonquery.New(node).StrWithDefault("next_page_token", "")
	}
}

// Before parsing the records if applicable API will be called one time to get model describing custom fields.
// Object will then be enhanced using model.
func (c *Connector) parseReadRecords(
	ctx context.Context, config common.ReadParams, jsonPath string,
) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		arr, err := jsonquery.New(node).Array(jsonPath, true)
		if err != nil {
			return nil, err
		}

		customFields, err := c.requestCustomFields(ctx, config.ObjectName)
		if err != nil {
			return nil, err
		}

		result := make([]map[string]any, len(arr))

		for index, object := range arr {
			item, err := c.enhanceObjectWithCustomFieldNames(object, customFields)
			if err != nil {
				return nil, err
			}

			result[index] = item
		}

		return result, nil
	}
}

// In general this does the usual JSON parsing.
// However, those objects that contain "custom_fields" are processed as follows:
// * Locate custom fields in JSON read response.
// * Replace ids with human-readable names, which is provided as argument.
// * Place fields at the top level of the object.
func (c *Connector) enhanceObjectWithCustomFieldNames(
	node *ajson.Node,
	fields map[int]modelCustomField,
) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		// no custom fields, the object is ready as is.
		return object, nil
	}

	customFieldsResponse, err := jsonquery.ParseNode[readCustomFieldsResponse](node)
	if err != nil {
		return nil, err
	}

	// Replace identifiers with human-readable field names which were found by making a call to "/model".
	for _, field := range customFieldsResponse.CustomFields {
		if model, ok := fields[field.ID]; ok {
			object[model.FieldName] = field.Content
		}
	}

	delete(object, "custom_fields")

	return object, nil
}
