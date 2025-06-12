package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/keap/metadata"
	keapv1 "github.com/amp-labs/connectors/scripts/openapi/keap/metadata/v1"
	keapv2 "github.com/amp-labs/connectors/scripts/openapi/keap/metadata/v2"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range getObjects() {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field.Name: field.Name,
				}, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

// V2 objects that appear in V1 take precedence.
// The latest version is always favoured.
func getObjects() []metadatadef.Schema {
	registry := datautils.Map[string, metadatadef.Schema]{}

	for _, obj := range keapv1.Objects() {
		registry[obj.ObjectName] = obj
	}

	for _, obj := range keapv2.Objects() {
		registry[obj.ObjectName] = obj
	}

	return registry.Values()
}
