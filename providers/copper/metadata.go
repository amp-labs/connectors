package copper

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/copper/internal/metadata"
)

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	allExistingCustomFields, err := c.fetchCustomFields(ctx)
	if err != nil {
		return nil, err
	}

	data := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		objectMetadata, err := metadata.Schemas.SelectOne(c.Module(), objectName)
		if err != nil {
			data.AppendError(objectName, err)

			continue
		}

		for _, field := range allExistingCustomFields.FilterByObjectName(objectName) {
			objectMetadata.AddFieldMetadata(field.Name(), common.FieldMetadata{
				DisplayName:  field.DisplayName,
				ValueType:    getFieldValueType(field),
				ProviderType: field.DataType,
				Values:       field.getValues(),
			})
		}

		data.Result[objectName] = *objectMetadata
	}

	return data, nil
}

func getFieldValueType(field customFieldResponse) common.ValueType {
	switch field.DataType {
	case "String", "Text", "URL":
		return common.ValueTypeString
	case "Currency", "Float", "Percentage":
		return common.ValueTypeFloat
	case "Checkbox":
		return common.ValueTypeBoolean
	case "Date":
		return common.ValueTypeDateTime
	case "Dropdown":
		return common.ValueTypeSingleSelect
	case "MultiSelect":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}
