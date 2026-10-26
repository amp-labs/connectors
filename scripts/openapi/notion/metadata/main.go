// Extracts Notion object schemas from the OpenAPI spec and writes
// providers/notion/metadata/schemas.json.
package main

import (
	"log/slog"
	"slices"
	"strings"

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
// Pages and data sources come only from POST /v1/search, which cannot take a time window.
// Blocks are listed at GET /v1/blocks/{block_id}/children, which needs a parent id.
// Comments and views need a parent id. Search is an action.
// None of those are connector objects.
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

// Component overlays for the list objects above.
// The list extractor leaves oneOf properties without a type and does not collect const.
//
//nolint:gochecknoglobals
var resourceObjects = []resourceObject{
	{name: "users", displayName: "Users", component: "userObjectResponse", path: "/users"},
	{name: "file_uploads", displayName: "File Uploads", component: "fileUploadObjectResponse", path: "/file_uploads"},
}

type resourceObject struct {
	name        string
	displayName string
	component   string
	path        string
}

func main() {
	schemas, registry := schemasFromLists()

	doc, err := files.Document()
	goutils.MustBeNil(err)

	addResourceObjects(schemas, doc)

	goutils.MustBeNil(outputMetadata.FlushSchemas(schemas))
	goutils.MustBeNil(outputMetadata.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "listObjects", len(supportedObjects), "resourceObjects", len(resourceObjects))
}

func schemasFromLists() (
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, any],
	datautils.NamedLists[string],
) {
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

		addObject(
			schemas,
			object.ObjectName,
			object.DisplayName,
			pathWithoutVersion(object.URLPath),
			object.ResponseKey,
			object.Fields,
		)

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	return schemas, registry
}

func addResourceObjects(
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any],
	doc *openapi3.T,
) {
	for _, resource := range resourceObjects {
		component := doc.Components.Schemas[resource.component]
		if component == nil || component.Value == nil {
			slog.Error("component schema missing", "component", resource.component)

			continue
		}

		addObject(
			schemas,
			resource.name,
			resource.displayName,
			resource.path,
			"results",
			fieldsFromSchema(component.Value),
		)
	}
}

// pathWithoutVersion drops the API version. Object paths are the resource only,
// for example OpenAPI /v1/users becomes /users.
func pathWithoutVersion(path string) string {
	trimmed, ok := strings.CutPrefix(path, "/v1")
	if !ok || trimmed == "" {
		return path
	}

	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}

	return trimmed
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

// walkSchema flattens a component into one field map.
// It follows allOf, oneOf, and anyOf, then records each direct property.
// Properties that share a name are merged. const and enum on those properties
// are unioned. The first non-null type is kept. Nested object properties are
// not expanded.
//
// Notion often has no enum array. Each oneOf or anyOf branch is one option,
// and that option is a const on the branch.
//
// userObjectResponse:
//
//	allOf:
//	  userObjectResponseCommon
//	    object: const "user"
//	  oneOf:
//	    personUserObjectResponse
//	      type: const "person"
//	    botUserObjectResponse
//	      type: const "bot"
//
// Collected fields: object ["user"], type ["bot", "person"].
//
// A real enum is collected from the property itself.
// fileUploadObjectResponse.status is enum ["pending", "uploaded", "expired", "failed"].
// filename is oneOf [string, null], so the collected type is string and there are no options.
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
