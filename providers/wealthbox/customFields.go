package wealthbox

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// requestCustomFields fetches the custom field definitions for a given object via
// GET /v1/categories/custom_fields?document_type=X.
// https://dev.wealthbox.com/#topics-custom-fields
//
// Objects without a document_type mapping (e.g. notes) don't support custom fields
// and an empty result is returned.
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]customFieldDefinition, error) {
	documentType, ok := documentTypeByObjectName[objectName]
	if !ok {
		return map[string]customFieldDefinition{}, nil
	}

	url, err := c.getCustomFieldsURL(documentType)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	response, err := common.UnmarshalJSON[customFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if response == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	fields := make(map[string]customFieldDefinition, len(response.CustomFields))
	for _, field := range response.CustomFields {
		fields[field.Name] = field
	}

	return fields, nil
}

// nolint:tagliatelle
type customFieldsResponse struct {
	CustomFields []customFieldDefinition `json:"custom_fields"`
}

// nolint:tagliatelle
type customFieldDefinition struct {
	ID           int                 `json:"id"`
	Name         string              `json:"name"`
	DocumentType string              `json:"document_type"`
	FieldType    string              `json:"field_type"`
	Options      []customFieldOption `json:"options"`
}

type customFieldOption struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

// https://dev.wealthbox.com/#topics-custom-fields
// Wealthbox does not publish an exhaustive list of field_type values;
// this covers the types observed in the docs and typical CRM custom field kinds.
func (f customFieldDefinition) getValueType() common.ValueType {
	switch f.FieldType {
	case "text", "textarea", "url", "email", "phone":
		return common.ValueTypeString
	case "number", "currency", "percentage":
		return common.ValueTypeFloat
	case "date", "datetime":
		return common.ValueTypeDate
	case "boolean":
		return common.ValueTypeBoolean
	case "single_select":
		return common.ValueTypeSingleSelect
	case "multi_select":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}

func (f customFieldDefinition) getValues() common.FieldValues {
	return datautils.ForEach(f.Options, func(option customFieldOption) common.FieldValue {
		return common.FieldValue{
			Value:        option.Label,
			DisplayValue: option.Label,
		}
	})
}
