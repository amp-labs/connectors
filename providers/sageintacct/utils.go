package sageintacct

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func mapSageIntacctTypeToValueType(sageType string) common.ValueType {
	switch sageType {
	case "string":
		return common.ValueTypeString
	case "integer", "number":
		return common.ValueTypeFloat
	case "boolean":
		return common.ValueTypeBoolean
	case "date", "date-time":
		return common.ValueTypeString
	default:
		return common.ValueTypeOther
	}
}

func mapValuesFromEnum(fieldDef SageIntacctFieldDef) []common.FieldValue {
	values := []common.FieldValue{}

	if len(fieldDef.Enum) > 0 {
		for _, v := range fieldDef.Enum {
			values = append(values, common.FieldValue{
				DisplayValue: naming.CapitalizeFirstLetter(v),
				Value:        v,
			})
		}
	}

	return values
}
