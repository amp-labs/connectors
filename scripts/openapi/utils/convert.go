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
			ValueType:    GetFieldValueType(field),
			ProviderType: field.Type,
			Values:       GetFieldValueOptions(field),
		},
	}
}

func GetFieldValueType(field metadatadef.Field) common.ValueType {
	switch field.Type {
	case "integer":
		return common.ValueTypeInt
	case "boolean":
		return common.ValueTypeBoolean
	case "string":
		if len(field.ValueOptions) != 0 {
			return common.ValueTypeSingleSelect
		}

		return common.ValueTypeString
	default:
		// object, array
		return common.ValueTypeOther
	}
}

func GetFieldValueOptions(field metadatadef.Field) staticschema.FieldValues {
	if len(field.ValueOptions) == 0 {
		return nil
	}

	values := make(staticschema.FieldValues, len(field.ValueOptions))
	for index, option := range field.ValueOptions {
		values[index] = staticschema.FieldValue{
			Value:        option,
			DisplayValue: option,
		}
	}

	return values
}
