package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"

	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/metadata/user"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}
	lists := datautils.IndexedLists[common.ModuleID, metadatadef.Schema]{}

	lists.Add(zoom.Users, user.Objects()...)

	for module, objects := range lists {
		for _, object := range objects {
			if object.Problem != nil {
				slog.Error("schema not extracted",
					"objectName", object.ObjectName,
					"error", object.Problem,
				)
			}

			for _, field := range object.Fields {
				path := "users/v2" + object.URLPath
				schemas.Add(module, object.ObjectName, object.DisplayName, path, object.ResponseKey,
					staticschema.FieldMetadataMapV2{
						field.Name: staticschema.FieldMetadata{
							DisplayName:  fieldNameConvertToDisplayName(field.Name),
							ValueType:    providerTypeConvertToValueType(field.Type),
							ProviderType: field.Type,
							ReadOnly:     false,
							Values:       nil,
						},
					}, nil)
			}

			for _, queryParam := range object.QueryParams {
				registry.Add(queryParam, object.ObjectName)
			}
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")

}

func fieldNameConvertToDisplayName(fieldName string) string {
	return api3.CapitalizeFirstLetterEveryWord(
		api3.CamelCaseToSpaceSeparated(fieldName),
	)
}

func providerTypeConvertToValueType(providerType string) common.ValueType {
	switch providerType {
	case "integer":
		return common.ValueTypeInt
	case "string":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	default:
		// Ex: object, array
		return common.ValueTypeOther
	}
}
