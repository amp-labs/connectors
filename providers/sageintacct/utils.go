package sageintacct

import (
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/amp-common/jsonpath"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/goutils"
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

func buildReadBody(params common.ReadParams) (map[string]any, error) {
	fieldNames := params.Fields.List()

	// Fields names are in bracket notation format e.g. $['audit']['createdby']
	// we need to convert them to dot notation e.g. audit.createdby for the API request
	dotNotation, err := convertFieldsToDotNotation(fieldNames)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"object":      params.ObjectName,
		"fields":      dotNotation,
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

// flattenFields flattens nested field definitions into JSONPath bracket notation.
// For example, audit.createdByUser.key becomes a single field entry.
func flattenFields(pathParts []string, fields map[string]SageIntacctFieldDef) map[string]common.FieldMetadata {
	result := make(map[string]common.FieldMetadata)

	for fieldName, fieldDef := range fields {
		parts := append(pathParts, fieldName)
		fullPath := jsonpath.ToNestedPath(parts...)

		result[fullPath] = common.FieldMetadata{
			DisplayName:  naming.CapitalizeFirstLetterEveryWord(fullPath),
			ValueType:    mapSageIntacctTypeToValueType(fieldDef.Type),
			ProviderType: fieldDef.Type,
			ReadOnly:     goutils.Pointer(fieldDef.ReadOnly),
			Values:       mapValuesFromEnum(fieldDef),
		}
	}

	return result
}

// flattenGroups flattens group definitions into JSONPath bracket notation.
// Groups can contain nested field definitions that are processed into flat paths.
func flattenGroups(pathParts []string, groups map[string]SageIntacctGroup) map[string]common.FieldMetadata {
	result := make(map[string]common.FieldMetadata)

	for groupName, group := range groups {
		parts := append(pathParts, groupName)
		groupFields := flattenFields(parts, group.Fields)
		maps.Copy(result, groupFields)
	}

	return result
}

// flattenRefs flattens reference (nested object) definitions into JSONPath bracket notation.
// Refs can contain both fields and nested groups, which are all processed into flat paths.
func flattenRefs(pathParts []string, refs map[string]SageIntacctRef) map[string]common.FieldMetadata {
	result := make(map[string]common.FieldMetadata)

	for refName, ref := range refs {
		parts := append(pathParts, refName)

		refFields := flattenFields(parts, ref.Fields)
		maps.Copy(result, refFields)

		if len(ref.Groups) > 0 {
			refGroups := flattenGroups(parts, ref.Groups)
			maps.Copy(result, refGroups)
		}
	}

	return result
}

func convertBracketToDotNotation(path string) (string, error) {
	parsedPathResult, err := jsonpath.ParsePath(path)
	if err != nil {
		return "", err
	}

	var dotNotationParts []string
	for _, segment := range parsedPathResult {
		dotNotationParts = append(dotNotationParts, segment.Key)
	}

	return strings.Join(dotNotationParts, "."), nil
}

// convertFieldsToDotNotation converts bracket notation fields to dot notation.
// Example: $['audit']['createdby'] -> audit.createdby
func convertFieldsToDotNotation(fieldNames []string) ([]string, error) {
	dotNotation := make([]string, 0, len(fieldNames))

	for _, field := range fieldNames {
		dotNotationField, err := convertBracketToDotNotation(field)
		if err != nil {
			return nil, err
		}
		dotNotation = append(dotNotation, dotNotationField)
	}

	return dotNotation, nil
}
