package main

import (
	"slices"
	"strings"

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
		Map(withoutCompanyID).
		Map(mapping.Fields(
			typedField,
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

	return typedScalar(field, options[0])
}

func typedScalar(field spec.Field, option *openapi3.Schema) spec.Field {
	switch {
	case option.Type.Is(openapi3.TypeString):
		field.Type = openapi3.TypeString
		field.ValueType, field.ValueOptions = stringValueType(option)
	case option.Type.Is(openapi3.TypeNumber) && isIdentifier(field.Name):
		field.Type = openapi3.TypeNumber
		field.ValueType = common.ValueTypeInt
	case option.Type.Is(openapi3.TypeNumber):
		field.Type = openapi3.TypeNumber
		field.ValueType = common.ValueTypeFloat
	case option.Type.Is(openapi3.TypeInteger):
		field.Type = openapi3.TypeInteger
		field.ValueType = common.ValueTypeInt
	case option.Type.Is(openapi3.TypeBoolean):
		field.Type = openapi3.TypeBoolean
		field.ValueType = common.ValueTypeBoolean
	case option.Type.Is(openapi3.TypeObject):
		field.Type = openapi3.TypeObject
		field.ValueType = common.ValueTypeOther
	}

	return field
}

// isIdentifier tells the ids the spec declares as `number` (`id`, `companyId`): they are whole
// numbers, while other numbers such as a meeting's `duration` in seconds have fractions.
func isIdentifier(name string) bool {
	return name == "id" || strings.HasSuffix(name, "Id")
}

func stringValueType(option *openapi3.Schema) (common.ValueType, []string) {
	switch {
	case option.Format == "date-time":
		return common.ValueTypeDateTime, nil
	case len(option.Enum) != 0:
		return common.ValueTypeSingleSelect, enumValues(option.Enum)
	default:
		return common.ValueTypeString, nil
	}
}

// withoutCompanyID drops the optional companies `id`: GET /companies never returns it, and a company
// is identified by its `domain` (GET /company/{domain}). For the same reason people's `companyId`
// is not typed as a reference: no companies record carries the id it would point to.
func withoutCompanyID(schema spec.Schema) spec.Schema {
	if schema.UrlPath == "/companies" {
		delete(schema.Fields, "id")
	}

	return schema
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
