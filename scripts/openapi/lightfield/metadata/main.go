package main

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/map/mappingfields"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/lightfield/internal/files"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	extractor, err := files.OpenApiFile.Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		api.List.WithArrayLocator(api.ArrayLocationAtData),
		api.List.WithPropertyFlattener(func(objectName, fieldName string) bool {
			// Read promotes every entry of these maps to a top-level field,
			// keyed exactly as Lightfield keys them ($name, $owner, custom slugs).
			return fieldName == "fields" || fieldName == "relationships"
		}),
	)
	goutils.MustBeNil(err)

	// Only the collections this connector reads.
	readObjects := filters.NewRules(
		"/v1/accounts",
		"/v1/contacts",
		"/v1/opportunities",
		"/v1/members",
	)

	pipe := pipeline.NewSchemaPipe(getSchemas).
		Filter(filters.KeepWithPath(filters.AndPathMatcher{
			filters.NoIdPath{}, filters.NewAllowPathStrategy(readObjects),
		})).
		Reduce(reducers.ShortestNameFromUrlWithFullFallback).
		Map(mapping.DisplayNameFromObjectName).
		Map(mapping.DisplayNameFormat(
			mapping.CamelCaseToSpaceSeparated,
			mapping.CapitalizeFirstLetterEveryWord,
		)).
		Map(mapping.Fields(
			typedValueField,
			timestampField,
			mappingfields.DisplayNameFormat(func(_, name string) string {
				return mapping.CapitalizeFirstLetterEveryWord(
					mapping.CamelCaseToSpaceSeparated(strings.TrimPrefix(name, "$")))
			}),
		))

	output.PrintUnusedRules("readObjects", readObjects)
	goutils.MustBeNil(output.WriteMetadata(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
}

// typedValueField describes a flattened `{value, valueType}` field by its inner value,
// which is what Read returns, and keeps Lightfield's valueType as the provider type.
// A flattened relationship `{cardinality, objectType, values}` is described by its cardinality;
// Read returns its `values` array.
func typedValueField(field spec.Field) spec.Field {
	if field.Origin.Schema == nil {
		return field
	}

	property, ok := field.Origin.Schema.Properties[field.Name]
	if !ok || property.Value == nil {
		return field
	}

	if valueType := singleEnum(property.Value.Properties["valueType"]); valueType != "" {
		field.Type = valueType
		field.ValueType = valueTypeOf(property.Value.Properties["value"])

		return field
	}

	if cardinality := singleEnum(property.Value.Properties["cardinality"]); cardinality != "" {
		field.Type = cardinality
		field.ValueType = relationshipValueType(cardinality)

		return field
	}

	field.ValueType = valueTypeOf(property)

	return field
}

// relationshipValueType matches the connector's typing of relationships: HAS_ONE is a lookup,
// HAS_MANY is a list of IDs and stays other. The only spec relationship among the read objects is
// members.$user, whose target `user` is not an object this connector reads, so it has no ReferenceTo.
func relationshipValueType(cardinality string) common.ValueType {
	if cardinality == "HAS_ONE" {
		return common.ValueTypeReference
	}

	return common.ValueTypeOther
}

// timestampField marks the record timestamps, documented as ISO 8601 strings, as datetimes.
func timestampField(field spec.Field) spec.Field {
	if field.Name == "createdAt" || field.Name == "updatedAt" {
		field.ValueType = common.ValueTypeDateTime
	}

	return field
}

func singleEnum(schema *openapi3.SchemaRef) string {
	if schema == nil || schema.Value == nil || len(schema.Value.Enum) != 1 {
		return ""
	}

	value, _ := schema.Value.Enum[0].(string)

	return value
}

// valueTypeOf maps the schema of a field's inner value to an Ampersand type.
// A nullable string (anyOf string|null) is still a string; anything else is other.
func valueTypeOf(schema *openapi3.SchemaRef) common.ValueType {
	if schema == nil || schema.Value == nil {
		return common.ValueTypeOther
	}

	types := schemaTypes(schema.Value)
	nonNull := make([]string, 0, len(types))

	for _, typeName := range types {
		if typeName != openapi3.TypeNull {
			nonNull = append(nonNull, typeName)
		}
	}

	if len(nonNull) == 1 && nonNull[0] == openapi3.TypeString {
		return common.ValueTypeString
	}

	return common.ValueTypeOther
}

// schemaTypes collects the schema's own types and those of each anyOf option.
func schemaTypes(schema *openapi3.Schema) []string {
	types := make([]string, 0)
	if schema.Type != nil {
		types = append(types, schema.Type.Slice()...)
	}

	for _, option := range schema.AnyOf {
		if option.Value != nil && option.Value.Type != nil {
			types = append(types, option.Value.Type.Slice()...)
		}
	}

	return types
}
