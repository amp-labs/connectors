// Extracts Notion object schemas from the OpenAPI spec and writes
// providers/notion/metadata/schemas.json.
package main

import (
	"log/slog"
	"slices"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/notion/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi3"
)

//nolint:gochecknoglobals
var outputMetadata = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
	fileconv.NewPath("providers/notion/metadata"),
)

// Workspace collections with a GET list and no parent id.
// Pages, data sources, and blocks are added from their object schemas below.
// Comments and views need a parent id, so they are not connector objects.
// Search is an action, not an object.
//
//nolint:gochecknoglobals
var supportedObjects = map[string]string{
	"/v1/users":        "users",
	"/v1/file_uploads": "file_uploads",
}

//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	"users":        "Users",
	"file_uploads": "File Uploads",
}

//nolint:gochecknoglobals
var objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
	"users":        "results",
	"file_uploads": "results",
},
	func(string) string {
		return ""
	},
)

// Resource objects whose list call is not a parent-free GET.
// Pages and data sources are listed with POST /v1/search.
// Blocks are listed with GET /v1/blocks/{block_id}/children.
// users and file_uploads are also listed here so nullable fields keep their types.
// The OpenAPI list extractor leaves oneOf properties without a type.
//
//nolint:gochecknoglobals
var resourceObjects = []resourceObject{
	{name: "users", displayName: "Users", component: "userObjectResponse", path: "/v1/users"},
	{name: "file_uploads", displayName: "File Uploads", component: "fileUploadObjectResponse", path: "/v1/file_uploads"},
	{name: "pages", displayName: "Pages", component: "pageObjectResponse", path: "/v1/pages"},
	{name: "data_sources", displayName: "Data Sources", component: "dataSourceObjectResponse", path: "/v1/data_sources"},
	{name: "blocks", displayName: "Blocks", component: "blockObjectResponse", path: "/v1/blocks"},
}

type resourceObject struct {
	name        string
	displayName string
	component   string
	path        string
}

func main() {
	explorer, err := files.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	allowPaths := make([]string, 0, len(supportedObjects))
	for path := range supportedObjects {
		allowPaths = append(allowPaths, path)
	}

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowPaths),
		supportedObjects,
		displayNameOverrides,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		addObject(schemas, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey, object.Fields)

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	doc, err := files.Document()
	goutils.MustBeNil(err)

	for _, resource := range resourceObjects {
		component := doc.Components.Schemas[resource.component]
		if component == nil || component.Value == nil {
			slog.Error("component schema missing", "component", resource.component)

			continue
		}

		addObject(schemas, resource.name, resource.displayName, resource.path, "results", fieldsFromSchema(component.Value))
	}

	goutils.MustBeNil(outputMetadata.FlushSchemas(schemas))
	goutils.MustBeNil(outputMetadata.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "listObjects", len(supportedObjects), "resourceObjects", len(resourceObjects))
}

func addObject(
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any],
	name, displayName, path, responseKey string,
	fields metadatadef.Fields,
) {
	for _, field := range fields {
		schemas.Add("", name, displayName, path, responseKey,
			utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, nil)
	}
}

type collectedField struct {
	typ   string
	enums map[string]struct{}
}

func fieldsFromSchema(schema *openapi3.Schema) metadatadef.Fields {
	collected := map[string]*collectedField{}
	walkSchema(schema, collected)

	fields := make(metadatadef.Fields, len(collected))

	for name, field := range collected {
		options := make([]string, 0, len(field.enums))
		for option := range field.enums {
			options = append(options, option)
		}

		slices.Sort(options)

		fields[name] = metadatadef.Field{
			Name:         name,
			Type:         field.typ,
			ValueOptions: options,
		}
	}

	return fields
}

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
