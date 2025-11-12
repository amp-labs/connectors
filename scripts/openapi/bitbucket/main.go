package main

import (
	_ "embed"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints     []string              //nolint: gochecknoglobals
	displayNameOverride = map[string]string{} //nolint: gochecknoglobals
	objectEndpoints     = map[string]string{} //nolint: gochecknoglobals

	//go:embed swagger.json
	bitbucketAPI []byte

	InputBitBucket  = api3.NewOpenapiFileManager[any](bitbucketAPI)        //nolint: gochecknoglobals
	OutputBitBucket = scrapper.NewWriter[staticschema.FieldMetadataMapV2]( //nolint: gochecknoglobals
		fileconv.NewPath("scripts/openapi/bitbucket"))
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range Objects() {
		urlPath := object.URLPath
		objectName := object.ObjectName

		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", objectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			fieldMetadataMap := staticschema.FieldMetadataMapV2{
				field.Name: staticschema.FieldMetadata{
					DisplayName:  fieldNameConvertToDisplayName(field.Name),
					ValueType:    providerTypeConvertToValueType(field.Type),
					ProviderType: field.Type,
					Values:       nil,
				},
			}

			schemas.Add(common.ModuleRoot, objectName, object.DisplayName, urlPath,
				object.ResponseKey, fieldMetadataMap, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	goutils.MustBeNil(OutputBitBucket.FlushSchemas(schemas))
	goutils.MustBeNil(OutputBitBucket.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := InputBitBucket.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		func(objectName, fieldName string) bool {
			return true // never called
		},
	)
	goutils.MustBeNil(err)

	return objects
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
