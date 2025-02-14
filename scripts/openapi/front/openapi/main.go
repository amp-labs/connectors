package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/front/metadata"
	"github.com/amp-labs/connectors/providers/front/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/me",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/company/tags":                "company_tags",
		"/company/rules":               "company_rules",
		"/accounts/custom_fields":      "accounts_custom_fields",
		"/contacts/custom_fields":      "contacts_custom_fields",
		"/conversations/custom_fields": "conversations_custom_fields",
		"/inboxes/custom_fields":       "inboxes_custom_fields",
		"/links/custom_fields":         "links_custom_fields",
		"/teammates/custom_fields":     "teammates_custom_fields",
	}
)

func main() {
	explorerOption := api3.WithArrayItemAutoSelection()
	locator := func(objectName, fieldName string) bool {
		return fieldName == "_results"
	}

	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithParameterFilterGetMethod(api3.OnlyOptionalQueryParameters),
		explorerOption,
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil, locator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field.Name: field.Name,
				}, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
