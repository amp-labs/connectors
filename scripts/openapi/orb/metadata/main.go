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

	// Only the collections this connector reads. Spec paths are relative to
	// the server https://api.withorb.com/v1; the connector adds the version.
	readObjects := filters.NewRules(
		"/customers",
		"/invoices",
		"/metrics",
		"/subscriptions",
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
			mappingfields.DisplayNameFormat(func(_, name string) string {
				return mapping.CapitalizeFirstLetterEveryWord(strings.ReplaceAll(name, "_", " "))
			}),
		))

	output.PrintUnusedRules("readObjects", readObjects)
	goutils.MustBeNil(output.WriteMetadata(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
}

// typedField types a field by its property schema. Orb declares most optional fields
// as `oneOf: [<schema>, {type: null}]`, which the extractor leaves untyped; the
// non-null option decides the type here. Timestamps (`format: date-time`) and dates
// (`format: date`) are typed as such, and string enums keep their options.
func typedField(field spec.Field) spec.Field {
	if field.Origin.Schema == nil {
		return field
	}

	property, ok := field.Origin.Schema.Properties[field.Name]
	if !ok || property.Value == nil {
		return field
	}

	schema := nonNullOption(property.Value)
	if schema == nil || schema.Type == nil || len(schema.Type.Slice()) != 1 {
		return field
	}

	field.Type = schema.Type.Slice()[0]

	if field.Type != openapi3.TypeString {
		return field
	}

	switch schema.Format {
	case "date-time":
		field.ValueType = common.ValueTypeDateTime
	case "date":
		field.ValueType = common.ValueTypeDate
	}

	if len(field.ValueOptions) == 0 {
		for _, option := range schema.Enum {
			if value, ok := option.(string); ok {
				field.ValueOptions = append(field.ValueOptions, value)
			}
		}
	}

	return field
}

// nonNullOption returns the schema itself, or the only non-null option of a oneOf/anyOf.
func nonNullOption(schema *openapi3.Schema) *openapi3.Schema {
	options := append(append([]*openapi3.SchemaRef{}, schema.OneOf...), schema.AnyOf...)
	if len(options) == 0 {
		return schema
	}

	var found *openapi3.Schema

	for _, option := range options {
		if option == nil || option.Value == nil {
			return nil
		}

		if option.Value.Type != nil && option.Value.Type.Is(openapi3.TypeNull) {
			continue
		}

		if found != nil {
			return nil
		}

		found = option.Value
	}

	return found
}
