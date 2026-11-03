// Extracts Notion object schemas from the OpenAPI spec and writes
// providers/notion/metadata/objectsMetadata.json and endpoints.json.
package main

import (
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/map/mappingfields"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
	"github.com/amp-labs/connectors/scripts/openapi/notion/internal/files"
	"github.com/getkin/kin-openapi/openapi3"
)

// Workspace collections with a GET list and no parent id.
// Pages and data sources come only from POST /v1/search, which cannot take a time window.
// Blocks are listed at GET /v1/blocks/{block_id}/children, which needs a parent id.
// Comments and views need a parent id. Search is an action.
// None of those are connector objects.
//
//nolint:gochecknoglobals
var readObjects = filters.NewRules(
	"/v1/users",
	"/v1/file_uploads",
)

//nolint:gochecknoglobals
var displayNameOverride = core.NewTrackedMap(map[string]string{
	"/users":        "Users",
	"/file_uploads": "File Uploads",
})

//nolint:gochecknoglobals
var objectNameByPath = map[string]string{
	"/v1/users":        "users",
	"/v1/file_uploads": "file_uploads",
}

// Component overlays for the list objects above.
// The list extractor leaves oneOf properties without a type and does not collect const.
//
//nolint:gochecknoglobals
var componentByObject = map[string]string{
	"users":        "userObjectResponse",
	"file_uploads": "fileUploadObjectResponse",
}

func main() {
	extractor, err := files.OpenApiFile.Extractor()
	goutils.MustBeNil(err)

	doc, err := files.Document()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		api.List.WithArrayLocator(func(_, fieldName string) bool {
			return fieldName == "results"
		}),
	)
	goutils.MustBeNil(err)

	pipe := pipeline.NewSchemaPipe(getSchemas).
		Filter(filters.KeepWithPath(filters.AndPathMatcher{
			filters.NoIdPath{}, filters.NewAllowPathStrategy(readObjects),
		})).
		Map(setObjectName).
		Map(pathWithoutVersion).
		Map(mapping.DisplayNameOverride(displayNameOverride)).
		Map(applyComponentFields(doc)).
		Map(mapping.Fields(
			mappingfields.DisplayNameFormat(func(_, name string) string {
				return name
			}),
		))

	output.PrintUnusedRules("readObjects", readObjects)
	output.PrintRedundantKeys("displayNameOverride", displayNameOverride)

	goutils.MustBeNil(output.WriteMetadata(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
}

func setObjectName(schema spec.Schema) spec.Schema {
	if name, ok := objectNameByPath[schema.UrlPath]; ok {
		schema.ObjectName = name
	}

	return schema
}

// pathWithoutVersion drops the API version. Object paths are the resource only,
// for example OpenAPI /v1/users becomes /users.
func pathWithoutVersion(schema spec.Schema) spec.Schema {
	trimmed, ok := strings.CutPrefix(schema.UrlPath, "/v1")
	if !ok || trimmed == "" {
		return schema
	}

	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}

	schema.UrlPath = trimmed

	return schema
}

func applyComponentFields(doc *openapi3.T) func(spec.Schema) spec.Schema {
	return func(schema spec.Schema) spec.Schema {
		componentName, ok := componentByObject[schema.ObjectName]
		if !ok {
			return schema
		}

		component := doc.Components.Schemas[componentName]
		if component == nil || component.Value == nil {
			return schema
		}

		schema.Fields = fieldsFromComponent(component.Value)

		return schema
	}
}

type collectedField struct {
	typ   string
	enums map[string]struct{}
}

func fieldsFromComponent(schema *openapi3.Schema) spec.Fields {
	collected := map[string]*collectedField{}
	walkSchema(schema, collected)

	fields := make(spec.Fields, len(collected))

	for name, field := range collected {
		options := make([]string, 0, len(field.enums))
		for option := range field.enums {
			options = append(options, option)
		}

		slices.Sort(options)

		specField := spec.Field{
			Name:         name,
			DisplayName:  name,
			Type:         field.typ,
			ValueOptions: options,
		}
		specField.ValueType = valueTypeForField(specField)

		fields[name] = specField
	}

	return fields
}

func valueTypeForField(field spec.Field) common.ValueType {
	switch field.Type {
	case openapi3.TypeBoolean:
		return common.ValueTypeBoolean
	case openapi3.TypeInteger:
		return common.ValueTypeInt
	case openapi3.TypeNumber:
		return common.ValueTypeFloat
	case openapi3.TypeString:
		if len(field.ValueOptions) != 0 {
			return common.ValueTypeSingleSelect
		}

		return common.ValueTypeString
	default:
		return common.ValueTypeOther
	}
}

// walkSchema flattens a component into one field map.
// It follows allOf, oneOf, and anyOf, then records each direct property.
// Properties that share a name are merged. const and enum on those properties
// are unioned. The first non-null type is kept. Nested object properties are
// not expanded.
//
// The shared list extractor copies an enum array into ValueOptions and does not
// walk oneOf. Notion often has no enum array; each oneOf or anyOf branch is one
// option, and that option is a const on the branch.
func walkSchema(schema *openapi3.Schema, collected map[string]*collectedField) {
	if schema == nil {
		return
	}

	for _, ref := range append(append(schema.AllOf, schema.OneOf...), schema.AnyOf...) {
		if ref != nil {
			walkSchema(ref.Value, collected)
		}
	}

	for name, property := range schema.Properties {
		if property == nil || property.Value == nil {
			continue
		}

		field, ok := collected[name]
		if !ok {
			field = &collectedField{enums: map[string]struct{}{}}
			collected[name] = field
		}

		if field.typ == "" {
			field.typ = schemaType(property.Value)
		}

		collectOptions(property.Value, field.enums)
	}
}

func collectOptions(schema *openapi3.Schema, enums map[string]struct{}) {
	if schema == nil {
		return
	}

	for _, value := range schema.Enum {
		if option, ok := value.(string); ok && option != "" {
			enums[option] = struct{}{}
		}
	}

	if option, ok := schema.Const.(string); ok && option != "" {
		enums[option] = struct{}{}
	}
}

func schemaType(schema *openapi3.Schema) string {
	if schema == nil {
		return ""
	}

	if typ := declaredType(schema); typ != "" {
		return typ
	}

	if typ := compositeType(schema); typ != "" {
		return typ
	}

	if schema.Items != nil {
		return "array"
	}

	if len(schema.Properties) > 0 {
		return "object"
	}

	return ""
}

func declaredType(schema *openapi3.Schema) string {
	if schema.Type == nil {
		return ""
	}

	for _, typ := range schema.Type.Slice() {
		if typ != "" && typ != "null" {
			return typ
		}
	}

	return ""
}

func compositeType(schema *openapi3.Schema) string {
	for _, ref := range append(append(schema.AllOf, schema.OneOf...), schema.AnyOf...) {
		if ref == nil {
			continue
		}

		if typ := schemaType(ref.Value); typ != "" {
			return typ
		}
	}

	return ""
}
