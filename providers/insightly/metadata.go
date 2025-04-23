package insightly

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/insightly/metadata"
)

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	data := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		objectMetadata, err := c.getObjectMetadata(ctx, objectName)
		if err != nil {
			data.AppendError(objectName, err)

			continue
		}

		fields, err := c.requestCustomFields(ctx, objectName)
		if err != nil {
			data.AppendError(objectName, err)
		}

		for fieldName, field := range fields {
			objectMetadata.AddFieldMetadata(fieldName, common.FieldMetadata{
				DisplayName:  field.DisplayName,
				ValueType:    getFieldValueType(field),
				ProviderType: field.Type,
				ReadOnly:     !field.Editable,
				Values:       getDefaultValues(field),
			})
		}

		data.Result[objectName] = *objectMetadata
	}

	return data, nil
}

func (c *Connector) getObjectMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	objectData, err := metadata.Schemas.SelectOne(c.Module(), objectName)
	if err != nil {
		// Check if the object is custom, then we still can create metadata.
		if errors.Is(err, common.ErrObjectNotSupported) && strings.HasSuffix(objectName, customMarker) {
			displayName, err := c.fetchCustomObjectDisplayName(ctx, objectName)
			if err != nil {
				return nil, err
			}

			// Custom object schema has the same format across every object type.
			fields := datautils.Map[string, common.FieldMetadata](
				customObjectSchema,
			).ShallowCopy()

			return common.NewObjectMetadata(displayName, fields), nil
		}

		return nil, err
	}

	return objectData, nil
}

func getFieldValueType(field customFieldResponse) common.ValueType {
	switch field.Type {
	case "NUMERIC", "PERCENT":
		return common.ValueTypeFloat
	case "TEXT", "MULTILINETEXT":
		return common.ValueTypeString
	case "MULTISELECT", "record-reference", "domain":
		return common.ValueTypeMultiSelect
	case "DROPDOWN":
		return common.ValueTypeSingleSelect
	case "DATE":
		return common.ValueTypeDate
	case "DATETIME":
		return common.ValueTypeDateTime
	default:
		// BIT, ARRAY, AUTONUMBER
		return common.ValueTypeOther
	}
}

func getDefaultValues(field customFieldResponse) common.FieldValues {
	if len(field.Options) == 0 {
		return nil
	}

	fields := make(common.FieldValues, len(field.Options))

	for index, option := range field.Options {
		fields[index] = common.FieldValue{
			Value:        strconv.Itoa(option.ID),
			DisplayValue: option.Value,
		}
	}

	return fields
}
