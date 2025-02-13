package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/gong/metadata"
	"github.com/amp-labs/connectors/providers/gong/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/me",

		// I am guessing that custom field processing should be part of the deep connector
		// implementation. Its possible that we would surface made up objects.
		// Ex: /contacts/custom_fields => contacts_custom_fields object
		// Ex: /inboxes/custom_fields => inboxes_custom_fields object
		// Ignoring all endpoints concerning custom fields:
		"*/custom_fields",

		// You should go over docs and see if we need "rules" objects alongside "company_rules".
		// Ignoring for now. If needed search for `objectEndpoints` variable for more info.
		"/company/rules",
		"/company/tags",
	}
	objectEndpoints = map[string]string{
		// TODO maybe map company/tags to company_tags
	}
)

func main() {
	// Use either `explorerOption` or `locator`.
	// Locator is intended when response has more than one field holding an array to resolve ambiguity.
	// (Feel free to suggest naming improvements.)
	// It seems locator is never called (given you pass `explorerOption`) because all responses have at max one array
	// in the response, which is stored under the same "_results" key.
	//
	// In other words even if array would be called differently `explorerOption` would pick that, and you will see
	// correct field at schema.json/modules/objects/<objectName>/responseKey.
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
