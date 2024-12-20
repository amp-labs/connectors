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
		// Display names will be coming from OpenAPI file and then space separated with every first letter capitalized.
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		// Read endpoints with required query parameters will be excluded from the search.
		api3.WithParameterFilterGetMethod(
			api3.OnlyOptionalQueryParameters,
		),
		// If response body has exactly one array the schema under that field will be chosen to describe our "Object".
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	// GET endpoints are used with some being ignored.
	// In case the response would ever have more than one array in its response the log would warn calling for action.
	// Either, the endpoint is to be ignored or ambiguity must be solved and explicitly tell which field holds
	// the schema of interest.
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
