package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/aweber/metadata"
	"github.com/amp-labs/connectors/providers/aweber/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		"/accounts/{accountId}",
		"/accounts/{accountId}/lists/{listId}/broadcasts/*",
		"/accounts/{accountId}/lists/{listId}/campaigns/*",
		"/accounts/{accountId}/lists/{listId}/campaigns/{campaignType}{campaignId}/links/*",
		"/accounts/{accountId}/lists/{listId}/custom_fields/*",
		"/accounts/{accountId}/integrations/*",
		"/accounts/{accountId}/lists/{listId}/landing_pages/*",
		"/accounts/{accountId}/lists/{listId}",
		"/accounts/{accountId}/lists?ws.op=find",
		"/accounts/{accountId}/lists/{listId}/segments/*",
		"/accounts/{accountId}/lists/{listId}/subscribers/*",
		"/accounts/{accountId}/lists/{listId}/purchases",
		"/accounts/{accountId}/lists/{listId}/web_form_split_tests/*",
		"/accounts/{accountId}/lists/{listId}/web_forms/*",
		"*?ws.op=*",
	}
	objectEndpoints = map[string]string{
		"/accounts/{accountId}/lists/{listId}/broadcasts":                                 "broadcasts",
		"/accounts/{accountId}/lists/{listId}/campaigns":                                  "campaigns",
		"/accounts/{accountId}/lists/{listId}/campaigns/{campaignType}{campaignId}/links": "campaign-links",
		"/accounts/{accountId}/lists/{listId}/custom_fields":                              "custom-fields",
		"/accounts/{accountId}/integrations":                                              "integrations",
		"/accounts/{accountId}/lists/{listId}/landing_pages":                              "landing-pages",
		"/accounts/{accountId}/lists":                                                     "lists",
		"/accounts/{accountId}/lists/{listId}/tags":                                       "tags",
		"/accounts/{accountId}/lists/{listId}/segments":                                   "segments",
		"/accounts/{accountId}/lists/{listId}/subscribers":                                "subscribers",
		"/accounts/{accountId}/lists/{listId}/web_forms":                                  "web-forms",
		"/accounts/{accountId}/lists/{listId}/web_form_split_tests":                       "web-form-split-tests",
	}
)

func main() {
	schemas := staticschema.NewMetadata()
	registry := datautils.NamedLists[string]{}
	lists := datautils.IndexedLists[common.ModuleID, api3.Schema]{}

	for _, manager := range openapi.FileManagers {
		// Every API section belongs to one module
		lists.Add("", ObjectsFromManager(manager)...)
	}

	for module, objects := range lists {
		for _, object := range objects {
			if object.Problem != nil {
				slog.Error("schema not extracted",
					"objectName", object.ObjectName,
					"error", object.Problem,
				)
			}

			for _, field := range object.Fields {
				schemas.Add(module, object.ObjectName, object.DisplayName,
					field, object.URLPath, object.ResponseKey, nil)
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

func ObjectsFromManager(manager *api3.OpenapiFileManager) []api3.Schema {
	explorer, err := manager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithPathIdentifiers(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		func(objectName, fieldName string) bool {
			return fieldName == "entries"
		},
	)
	goutils.MustBeNil(err)

	return objects
}
