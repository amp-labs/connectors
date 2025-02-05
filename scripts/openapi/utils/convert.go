package utilsopenapi

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func ConvertMetadataFieldToFieldMetadataMapV2(field metadatadef.Field) staticschema.FieldMetadataMapV2 {
	return staticschema.FieldMetadataMapV2{
		field.Name: staticschema.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    getFieldValueType(field),
			ProviderType: field.Type,
			ReadOnly:     false,
			Values:       getFieldValueOptions(field),
		},
	}
}

func getFieldValueType(field metadatadef.Field) common.ValueType {
	switch field.Type {
	case "integer":
		return common.ValueTypeInt
	case "boolean":
		return common.ValueTypeBoolean
	case "string":
		if len(field.EnumOptions) != 0 {
			return common.ValueTypeSingleSelect
		}

		return common.ValueTypeString
	default:
		// object, array
		return common.ValueTypeOther
	}
}

func getFieldValueOptions(field metadatadef.Field) staticschema.FieldValues {
	if len(field.EnumOptions) == 0 {
		return nil
	}

	values := make(staticschema.FieldValues, len(field.EnumOptions))
	for index, option := range field.EnumOptions {
		values[index] = staticschema.FieldValue{
			Value:        option,
			DisplayValue: option,
		}
	}

	return values
}
