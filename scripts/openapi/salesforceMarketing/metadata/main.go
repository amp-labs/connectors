package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/salesforceMarketing/metadata"
	"github.com/amp-labs/connectors/providers/salesforceMarketing/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	objectEndpoints = map[string]string{
		"/asset/v1/content/assets":        "assets",
		"/hub/v1/campaigns":               "campaigns",
		"/messaging/v1/email/definitions": "emailDefinitions",
		"/messaging/v1/sms/definitions":   "smsDefinitions",
	}

	displayNameOverride = map[string]string{
		"assets":           "Assets",
		"campaigns":        "Campaigns",
		"emailDefinitions": "Email Definitions",
		"smsDefinitions":   "SMS Definitions",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{
		"assets":           "items",
		"campaigns":        "items",
		"emailDefinitions": "definitions",
		"smsDefinitions":   "definitions",
	},
		func(objectName string) (fieldName string) {
			return ""
		},
	)

	// Define endpoints to ignore
	ignoreEndpoints = []string{
		// Add any endpoints that should be ignored
	}
)

func main() {

	explorer, err := openapi.ApiManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		// api3.WithParameterFilterGetMethod(
		// 	api3.OnlyOptionalQueryParameters,
		// ),
	)
	goutils.MustBeNil(err)

	// Read objects from the OpenAPI spec
	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNametoResponseField),
	)
	goutils.MustBeNil(err)

	// Create a new metadata object
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	// Process each object
	for _, object := range readObjects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
			continue
		}

		// Process fields
		for _, field := range object.Fields {
			// Add field to schema
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		// Process query parameters
		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")

}
