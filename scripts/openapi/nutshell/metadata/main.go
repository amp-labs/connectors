package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/nutshell/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Accounts, contacts, leads custom fields.
		"*/customfields/attributes",
	}
	displayNameOverride = map[string]string{
		"accounttypes":   "Account Types",
		"activitytypes":  "Activity Types",
		"competitormaps": "Competitor Maps",
		"productmaps":    "Product Maps",
		"stagesets":      "Stage Sets",
	}
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
			schemas.Add(common.ModuleRoot, objectName, object.DisplayName, urlPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	goutils.MustBeNil(files.OutputNutshell.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputNutshell.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputNutshell.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride,
		func(objectName, fieldName string) bool {
			if objectName == "competitormaps" {
				return fieldName == "competitorMaps"
			}

			slog.Warn("don't know how to find array property")

			return false
		},
	)
	goutils.MustBeNil(err)

	return objects
}
