package justcall

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
// List of object names which support custom fields.
// Docs on Custom fields: https://developer.justcall.io/reference/sd_list_contact_custom_fields_v21
var objectsWithCustomFields = datautils.NewStringSet("sales_dialer/contacts")

// requestCustomFields makes an API call to get model describing custom fields.
// For not applicable objects the empty mapping is returned.
// map[string]customFieldDefinition is a map of field label to the field definition itself.
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]customFieldDefinition, error) {
	if !objectsWithCustomFields.Has(objectName) {
		return map[string]customFieldDefinition{}, nil
	}

	url, err := c.getCustomFieldsURL()
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[customFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if fieldsResponse == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	fields := make(map[string]customFieldDefinition)
	for _, field := range fieldsResponse.Data {
		fields[field.Label] = field
	}

	return fields, nil
}

func (c *Connector) getCustomFieldsURL() (*urlbuilder.URL, error) {
	modulePath := metadata.Schemas.LookupModuleURLPath(c.ModuleID)

	return urlbuilder.New(c.BaseURL, modulePath, "/sales_dialer/contacts/custom-fields")
}

// customFieldsResponse represents the response from custom-fields endpoint.
// https://developer.justcall.io/reference/sd_list_contact_custom_fields_v21
type customFieldsResponse struct {
	Status string                  `json:"status"`
	Total  int                     `json:"total"`
	Data   []customFieldDefinition `json:"data"`
}

// customFieldDefinition represents a custom field definition.
type customFieldDefinition struct {
	Key   int    `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// getValueType maps JustCall field types to common.ValueType.
func (f customFieldDefinition) getValueType() common.ValueType {
	switch f.Type {
	case "string":
		return common.ValueTypeString
	case "number":
		return common.ValueTypeFloat
	case "date":
		return common.ValueTypeDate
	case "boolean":
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

// flattenCustomFields moves custom fields from the custom_fields array to the root level.
// This allows users to request custom fields by their label name.
func flattenCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsArray, err := jsonquery.New(node).ArrayOptional("custom_fields")
	if err != nil {
		return nil, err
	}

	if len(customFieldsArray) == 0 {
		return root, nil
	}

	// Custom fields are moved from the nested array to the root level.
	for _, customFieldObject := range customFieldsArray {
		field, err := jsonquery.ParseNode[readCustomField](customFieldObject)
		if err != nil {
			return nil, err
		}

		root[field.Label] = field.Value
	}

	return root, nil
}

// readCustomField represents a custom field as it appears in read responses.
//
// Example:
//
//	{
//	  "custom_fields": [
//	    {
//	      "key": 1101123,
//	      "label": "membership_status",
//	      "type": "string",
//	      "value": "member"
//	    }
//	  ]
//	}
type readCustomField struct {
	Key   int    `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Value string `json:"value"`
}
