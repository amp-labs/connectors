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
	Properties map[string]Property `yaml:"properties"`
}

type Property struct {
	Type        string `yaml:"type"`
	Format      string `yaml:"format"`
	Description string `yaml:"description"`
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

func constructProperties(data OpenAPI, schemas map[string]string) map[string][]string {
	properties := make(map[string][]string)

	for object, schemaRef := range schemas {
		schemaName := strings.Split(schemaRef, "/")[schemaIdx]
		if !strings.HasPrefix(schemaName, paginatedSchemaPrefix) {
			for property := range data.Components.Schemas[schemaName].Properties {
				properties[object] = append(properties[object], property)
			}
		} else {
			schemaName, _ := strings.CutPrefix(schemaName, paginatedSchemaPrefix)
			for property := range data.Components.Schemas[schemaName].Properties {
				properties[object] = append(properties[object], property)
			}
		}
	}

	for resourcePath, flds := range properties {
		// update the resourcePath to be the supportedObject.
		if len(strings.Split(resourcePath, "/")) == 4 { //nolint:gomnd,mnd
			properties[strings.Split(resourcePath, "/")[schemaIdx]] = flds
			delete(properties, resourcePath)
		}

		if len(strings.Split(resourcePath, "/")) > 4 { //nolint:gomnd,mnd
			objects := strings.Split(resourcePath, "/")
			customNamedObject := strings.Join(objects[3:], "_")
			properties[customNamedObject] = flds
			delete(properties, resourcePath)
		}
	}

	return properties
}

func saveSchemas(properties map[string][]string) {
	objectMetadata := make(map[string]staticschema.Object[staticschema.FieldMetadataMapV1])

	for object, flds := range properties {
		fldsMap := make(map[string]string)
		for _, fld := range flds {
			fldsMap[fld] = fld
		}

		om := staticschema.Object[staticschema.FieldMetadataMapV1]{
			DisplayName: object,
			Fields:      fldsMap,
		}
		objectMetadata[object] = om
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(&staticschema.Metadata[staticschema.FieldMetadataMapV1]{
		Modules: map[common.ModuleID]staticschema.Module[staticschema.FieldMetadataMapV1]{
			staticschema.RootModuleID: {
				Objects: objectMetadata,
			},
		},
	}))
}
