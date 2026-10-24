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
	"github.com/amp-labs/connectors/scripts/openapi/orb/internal/files"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	extractor, err := files.OpenApiFile.Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		api.List.WithArrayLocator(api.ArrayLocationAtData),
	)
	goutils.MustBeNil(err)

	// Only the collections this connector reads.
	readObjects := filters.NewRules(
		"/customers",
		"/subscriptions",
		"/invoices",
		"/metrics",
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
			nullableField,
			mappingfields.DisplayNameFormat(func(_, name string) string {
				return mapping.CapitalizeFirstLetterEveryWord(strings.ReplaceAll(name, "_", " "))
			}),
		))

	output.PrintUnusedRules("readObjects", readObjects)
	goutils.MustBeNil(output.WriteMetadata(files.OutputDir, pipe))
}

// nullableField types a field by its schema. Most Orb properties are OpenAPI 3.1 nullable unions,
// `oneOf: [{type: string, format: date-time}, {type: null}]`, which the extractor leaves untyped,
// so the field is typed by its non-null option. Timestamps are `format: date-time` strings.
func nullableField(field spec.Field) spec.Field {
	if field.Origin.Schema == nil {
		return field
	}

	property, ok := field.Origin.Schema.Properties[field.Name]
	if !ok || property.Value == nil {
		return field
	}

	schema := nonNullOption(property.Value)
	if schema == nil || schema.Type == nil || len(schema.Type.Slice()) == 0 {
		return field
	}

	field.Type = schema.Type.Slice()[0]
	field.ValueOptions = enumOptions(schema)
	field.ValueType = valueType(field.Type, schema.Format)

	return field
}

// valueType covers the types the output's default mapping doesn't tell apart: date-time and date
// strings, and numbers. An empty result keeps that default (string, integer, boolean, enum, other).
func valueType(typeName, format string) common.ValueType {
	switch {
	case typeName == openapi3.TypeString && format == "date-time":
		return common.ValueTypeDateTime
	case typeName == openapi3.TypeString && format == "date":
		return common.ValueTypeDate
	case typeName == openapi3.TypeNumber:
		return common.ValueTypeFloat
	default:
		return ""
	}
}

// nonNullOption returns the schema itself, or the only non-null option of its oneOf/anyOf union.
// A union of several non-null options has no single type and returns nil.
func nonNullOption(schema *openapi3.Schema) *openapi3.Schema {
	options := schema.OneOf
	if len(options) == 0 {
		options = schema.AnyOf
	}

	if len(options) == 0 {
		return schema
	}

	var found *openapi3.Schema

	for _, option := range options {
		if option.Value == nil || option.Value.Type.Is(openapi3.TypeNull) {
			continue
		}

		if found != nil {
			return nil
		}

		found = option.Value
	}

	return found
}

func enumOptions(schema *openapi3.Schema) []string {
	options := make([]string, 0, len(schema.Enum))

	for _, value := range schema.Enum {
		if option, ok := value.(string); ok {
			options = append(options, option)
		}
	}

	return options
}
