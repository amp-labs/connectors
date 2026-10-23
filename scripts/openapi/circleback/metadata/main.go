package main

import (
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/circleback/internal/files"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/map/mappingfields"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/getkin/kin-openapi/openapi3"
)

// Every list endpoint returns a bare JSON array of records.
func main() {
	extractor, err := files.OpenApiFile.Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET)
	goutils.MustBeNil(err)

	// Only the collections this connector reads.
	// /calendar/events (upcoming calendar schedule) and /search (a query over meetings) were not asked for.
	readObjects := filters.NewRules(
		"/meetings",
		"/action-items",
		"/tag",
		"/people",
		"/companies",
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
			typedField,
			referenceField,
			mappingfields.DisplayNameFormat(func(_, name string) string {
				return mapping.CapitalizeFirstLetterEveryWord(mapping.CamelCaseToSpaceSeparated(name))
			}),
		))

	output.PrintUnusedRules("readObjects", readObjects)
	goutils.MustBeNil(output.WriteMetadata(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
}

// typedField types scalar fields, including the nullable ones the spec declares as
// `anyOf: [{type: X}, {type: null}]`: a date-time string is a datetime, a number a float,
// and a string enum a single select. Arrays and objects keep the library's typing.
func typedField(field spec.Field) spec.Field {
	if field.Origin.Schema == nil {
		return field
	}

	property, ok := field.Origin.Schema.Properties[field.Name]
	if !ok || property.Value == nil {
		return field
	}

	options := nonNullOptions(property.Value)
	if len(options) != 1 {
		return field
	}

	option := options[0]

	switch {
	case option.Type.Is(openapi3.TypeString) && option.Format == "date-time":
		field.Type = openapi3.TypeString
		field.ValueType = common.ValueTypeDateTime
	case option.Type.Is(openapi3.TypeString) && len(option.Enum) != 0:
		field.Type = openapi3.TypeString
		field.ValueType = common.ValueTypeSingleSelect
		field.ValueOptions = enumValues(option.Enum)
	case option.Type.Is(openapi3.TypeString):
		field.Type = openapi3.TypeString
		field.ValueType = common.ValueTypeString
	case option.Type.Is(openapi3.TypeNumber):
		field.Type = openapi3.TypeNumber
		field.ValueType = common.ValueTypeFloat
	case option.Type.Is(openapi3.TypeInteger):
		field.Type = openapi3.TypeInteger
		field.ValueType = common.ValueTypeInt
	case option.Type.Is(openapi3.TypeBoolean):
		field.Type = openapi3.TypeBoolean
		field.ValueType = common.ValueTypeBoolean
	}

	return field
}

// referenceField links people to their company: `companyId` is the `id` of a companies record.
func referenceField(field spec.Field) spec.Field {
	if field.URLPath == "/people" && field.Name == "companyId" {
		field.ValueType = common.ValueTypeReference
		field.ReferenceTo = []string{"companies"}
	}

	return field
}

// nonNullOptions returns the schema itself, or its anyOf alternatives without the null one.
func nonNullOptions(schema *openapi3.Schema) []*openapi3.Schema {
	if len(schema.AnyOf) == 0 {
		return []*openapi3.Schema{schema}
	}

	options := make([]*openapi3.Schema, 0, len(schema.AnyOf))

	for _, ref := range schema.AnyOf {
		if ref.Value == nil || ref.Value.Type.Is(openapi3.TypeNull) {
			continue
		}

		options = append(options, ref.Value)
	}

	return options
}

func enumValues(enum []any) []string {
	values := make([]string, 0, len(enum))

	for _, value := range enum {
		if text, ok := value.(string); ok && !slices.Contains(values, text) {
			values = append(values, text)
		}
	}

	return values
}
