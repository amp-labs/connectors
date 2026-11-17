package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/chilipiper/metadata"
	"gopkg.in/yaml.v3"
)

const (
	schemaIdx             = 3
	paginatedSchemaPrefix = "PaginatedResult_"
	statusOK              = "200"
	contentType           = "application/json"
)

type OpenAPI struct {
	Paths      map[string]PathItem `yaml:"paths"`
	Components Components          `yaml:"components"`
}

type PathItem struct {
	Get Operation `yaml:"get,omitempty"`
}

type Operation struct {
	Responses map[string]Response `yaml:"responses"`
}

type Response struct {
	Content map[string]MediaType `yaml:"content,omitempty"`
}

type MediaType struct {
	Schema SchemaRef `yaml:"schema"`
}

type SchemaRef struct {
	Ref string `yaml:"$ref,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `yaml:"schemas"`
}

type Schema struct {
	Type       string              `yaml:"type"`
	Format     string              `yaml:"format"`
	Properties map[string]Property `yaml:"properties"`
}

type Property struct {
	Ref         string `yaml:"$ref"`
	Type        string `yaml:"type"`
	Format      string `yaml:"format"`
	Description string `yaml:"description"`
}

// objectNameOverrides renames objects derived from the URL path
// so that metadata uses the same object names as Read.
var objectNameOverrides = map[string]string{ // nolint:gochecknoglobals
	"workspace_users": "workspace/users",
}

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	file := "scripts/openapi/chilipiper/docs.yaml"

	d, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", file, err)
	}

	var data OpenAPI
	if err := yaml.Unmarshal(d, &data); err != nil {
		return err
	}

	objectSchemas := make(map[string]string)
	for path := range data.Paths {
		objectSchemas[path] = data.Paths[path].Get.Responses[statusOK].Content[contentType].Schema.Ref
		// delete the keys that have no schemas.
		if len(data.Paths[path].Get.Responses[statusOK].Content[contentType].Schema.Ref) == 0 {
			delete(objectSchemas, path)
		}
	}

	properties := constructProperties(data, objectSchemas)
	saveSchemas(properties)

	return nil
}

func constructProperties(data OpenAPI, schemas map[string]string) map[string]map[string]Property {
	properties := make(map[string]map[string]Property)

	for object, schemaRef := range schemas {
		schemaName := strings.Split(schemaRef, "/")[schemaIdx]
		schemaName = strings.TrimPrefix(schemaName, paginatedSchemaPrefix)

		// Schemas without properties produce no object.
		schemaProperties := data.Components.Schemas[schemaName].Properties
		if len(schemaProperties) == 0 {
			continue
		}

		properties[object] = make(map[string]Property)
		for name, property := range schemaProperties {
			properties[object][name] = resolveProperty(data, property)
		}
	}

	for resourcePath, flds := range properties {
		// update the resourcePath to be the supportedObject.
		if len(strings.Split(resourcePath, "/")) == 4 { //nolint:mnd
			properties[strings.Split(resourcePath, "/")[schemaIdx]] = flds
			delete(properties, resourcePath)
		}

		if len(strings.Split(resourcePath, "/")) > 4 { //nolint:mnd
			objects := strings.Split(resourcePath, "/")
			customNamedObject := strings.Join(objects[3:], "_")

			if override, ok := objectNameOverrides[customNamedObject]; ok {
				customNamedObject = override
			}

			properties[customNamedObject] = flds
			delete(properties, resourcePath)
		}
	}

	return properties
}

// resolveProperty follows a $ref to a component schema, so that the type of the referenced schema is used.
// Ex: "id" references "#/components/schemas/UserId", which is a string.
func resolveProperty(data OpenAPI, property Property) Property {
	if property.Ref == "" {
		return property
	}

	refSchema, ok := data.Components.Schemas[strings.Split(property.Ref, "/")[schemaIdx]]
	if !ok {
		return property
	}

	property.Type = refSchema.Type
	property.Format = refSchema.Format

	return property
}

// valueType maps OpenAPI types to Ampersand value types.
func valueType(property Property) common.ValueType {
	switch property.Type {
	case "string":
		switch property.Format {
		case "date-time":
			return common.ValueTypeDateTime
		case "date":
			return common.ValueTypeDate
		default:
			return common.ValueTypeString
		}
	case "boolean":
		return common.ValueTypeBoolean
	case "integer":
		return common.ValueTypeInt
	case "number":
		return common.ValueTypeFloat
	default:
		return common.ValueTypeOther
	}
}

func saveSchemas(properties map[string]map[string]Property) {
	objectMetadata := make(map[string]staticschema.Object[staticschema.FieldMetadataMapV2, any])

	for object, flds := range properties {
		fldsMap := make(staticschema.FieldMetadataMapV2)
		for name, property := range flds {
			fldsMap[name] = staticschema.FieldMetadata{
				DisplayName:  name,
				ValueType:    valueType(property),
				ProviderType: property.Type,
				ReadOnly:     nil,
				Values:       nil,
			}
		}

		om := staticschema.Object[staticschema.FieldMetadataMapV2, any]{
			DisplayName: object,
			Fields:      fldsMap,
		}
		objectMetadata[object] = om
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(&staticschema.Metadata[staticschema.FieldMetadataMapV2, any]{
		Modules: map[common.ModuleID]staticschema.Module[staticschema.FieldMetadataMapV2, any]{
			common.ModuleRoot: {
				Objects: objectMetadata,
			},
		},
	}))
}
