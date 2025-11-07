package sageintacct

import (
	"strconv"
	"time"

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

func buildReadBody(params common.ReadParams) (map[string]interface{}, error) {
	fieldNames := params.Fields.List()
	payload := map[string]any{
		"object":      params.ObjectName,
		"fields":      fieldNames,
		pageSizeParam: defaultPageSize,
		pageParam:     1,
	}

	if !objectNameNotSupportIncremental.Has(params.ObjectName) {
		dateFilters := make([]map[string]any, 0, 2) //nolint:mnd

		if !params.Since.IsZero() {
			dateFilters = append(dateFilters, map[string]any{
				"$gte": map[string]any{
					"audit.modifiedDateTime": params.Since.Format(time.RFC3339),
				},
			})
		}

		if !params.Until.IsZero() {
			dateFilters = append(dateFilters, map[string]any{
				"$lte": map[string]any{
					"audit.modifiedDateTime": params.Until.Format(time.RFC3339),
				},
			})
		}

		if len(dateFilters) > 0 {
			payload["filters"] = dateFilters
		}
	}

	if params.NextPage != "" {
		startPage, err := strconv.Atoi(string(params.NextPage))
		if err != nil {
			return nil, err
		}

		payload[pageParam] = startPage
	}

	return payload, nil
}
