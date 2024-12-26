package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/iterable/metadata"
	"github.com/amp-labs/connectors/providers/iterable/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// array of strings not objects
		"/api/users/forgotten",
		"/api/users/forgottenUserIds",
		// operations without GET method
		"*/upsert",
		"*/update",
		"*/cancel",
		"*/target",
		// singular object
		"/api/users/getFields",
		"/api/email/viewInBrowser",
		"/api/templates/email/get",
		"/api/users/byUserId",
		"/api/users/getByEmail",
		// requires query parameters
		"/api/campaigns/metrics",
		"/api/experiments/metrics",
		"/api/inApp/getMessages",
		"/api/templates/email/get",
		"/api/templates/inapp/get",
		"/api/templates/push/get",
		"/api/templates/sms/get",
		"/api/inApp/getPriorityMessage", // need email or userId
		"/api/users/getSentMessages",    // need email or userId
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithParameterFilterGetMethod(
			api3.OnlyOptionalQueryParameters,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil,
		arrayLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName,
				field, object.URLPath, object.ResponseKey, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func arrayLocator(objectName, fieldName string) bool {
	slog.Warn("unexpected call to locator, provider API was expected to have no ambiguous array fields",
		"object", objectName)

	return false
}
