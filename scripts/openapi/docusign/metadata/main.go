package main

import (
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/docusign/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	allowEndpoints = []string{
		"/v2.1/accounts/{accountId}/envelopes",
		"/v2.1/accounts/{accountId}/folders",
		"/v2.1/accounts/{accountId}/templates",
		"/v2.1/accounts/{accountId}/users",
		"/v2.1/accounts/{accountId}/bulk_send_batch",
		"/v2.1/accounts/{accountId}/bulk_send_lists",
		"/v2.1/accounts/{accountId}/users",
		"/v2.1/accounts/{accountId}/signing_groups",
		"/v2.1/accounts/{accountId}/tab_definitions",
	}

	overrideDisplayName = map[string]string{
		"bulk_send_batch": "Bulk Send Batch",
		"bulk_send_lists": "Bulk Send Lists",
		"signing_groups":  "Signing Groups",
		"tab_definitions": "Custom Tabs",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{
		"templates": "envelopeTemplates",
	},
		func(objectName string) (fieldName string) {
			return objectName
		})
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

	goutils.MustBeNil(files.OutputDocusignESignature.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputDocusignESignature.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputDocusignESignature.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(http.MethodGet,
		api3.NewAllowPathStrategy(allowEndpoints), nil, overrideDisplayName,
		api3.CustomMappingObjectCheck(objectNametoResponseField),
	)

	goutils.MustBeNil(err)

	return objects
}
