package main

import (
	"path"
	"slices"

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
	"github.com/amp-labs/connectors/scripts/openapi/sybill/internal/files"
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	conversationsPath = "/v1/conversations"
	// Response schema of GET /v1/conversations/{conversationId}, which Read calls for each conversation.
	conversationDetailSchema = "ConversationDetail"
)

func main() {
	extractor, err := files.OpenApiFile.Extractor()
	goutils.MustBeNil(err)

	doc, err := openapi3.NewLoader().LoadFromData(files.Data())
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		// Records are under the key named after the collection: {"conversations": [...], "pagination": {...}}.
		api.List.WithArrayLocator(func(urlPath, fieldName string) bool {
			return fieldName == path.Base(urlPath)
		}),
	)
	goutils.MustBeNil(err)

	// Only the collections this connector reads.
	readObjects := filters.NewRules(
		conversationsPath,
		"/v1/accounts",
		"/v1/deals",
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
		Map(conversationDetailFields(doc.Components.Schemas[conversationDetailSchema].Value)).
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

// conversationDetailFields adds the fields only the conversation detail returns (transcript, summary,
// recordings, category). Read fetches the detail of each listed conversation when any of them is
// requested, so they are readable fields of conversations.
func conversationDetailFields(detail *openapi3.Schema) pipeline.MapFunc[spec.Schema, spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		if schema.UrlPath != conversationsPath {
			return schema
		}

		for name := range detail.Properties {
			if _, ok := schema.Fields[name]; ok {
				continue
			}

			schema.Fields[name] = spec.Field{
				URLPath:     schema.UrlPath,
				Name:        name,
				DisplayName: name,
				Origin:      spec.FieldOrigin{ItemsSchema: detail, Schema: detail},
			}
		}

		return schema
	}
}

// typedField types a field by its property schema. Most properties are nullable, written as
// `anyOf: [T, null]`, which the extractor leaves untyped; the non-null options decide the type.
// Enums, including the two category enums of a conversation, become select values.
func typedField(field spec.Field) spec.Field {
	if field.Origin.Schema == nil {
		return field
	}

	property, ok := field.Origin.Schema.Properties[field.Name]
	if !ok || property.Value == nil {
		return field
	}

	options := nonNullOptions(property.Value)

	if values := enumValues(options); len(values) != 0 {
		field.Type = openapi3.TypeString
		field.ValueOptions = values
		field.ValueType = common.ValueTypeSingleSelect

		return field
	}

	if len(options) != 1 || options[0].Type == nil || len(options[0].Type.Slice()) == 0 {
		field.ValueType = common.ValueTypeOther

		return field
	}

	option := options[0]
	field.Type = option.Type.Slice()[0]
	field.ValueType = valueType(field.Type, option.Format)

	return field
}

func nonNullOptions(schema *openapi3.Schema) []*openapi3.Schema {
	if len(schema.AnyOf) == 0 {
		return []*openapi3.Schema{schema}
	}

	options := make([]*openapi3.Schema, 0, len(schema.AnyOf))

	for _, option := range schema.AnyOf {
		if option.Value == nil || option.Value.Type.Is(openapi3.TypeNull) {
			continue
		}

		options = append(options, option.Value)
	}

	return options
}

// enumValues joins the enums of every option; `other` is in both category enums and is kept once.
func enumValues(options []*openapi3.Schema) []string {
	values := make([]string, 0)

	for _, option := range options {
		for _, value := range option.Enum {
			if text, ok := value.(string); ok && !slices.Contains(values, text) {
				values = append(values, text)
			}
		}
	}

	return values
}

// Timestamps are ISO 8601 strings in UTC: https://api.sybill.ai/docs/data-models.html.
func valueType(typeName, format string) common.ValueType {
	switch typeName {
	case openapi3.TypeString:
		if format == "date-time" {
			return common.ValueTypeDateTime
		}

		return common.ValueTypeString
	case openapi3.TypeNumber:
		return common.ValueTypeFloat
	case openapi3.TypeInteger:
		return common.ValueTypeInt
	case openapi3.TypeBoolean:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}
