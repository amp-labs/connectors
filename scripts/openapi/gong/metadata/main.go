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
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	// Even though OpenAPI docs and official documentation say that some query parameters are required
	// in practice you still can make an API call without any specified.
	// Must include "calls" object.
	queryParamFilterExceptions = datautils.NewSet("calls") // nolint:gochecknoglobals

	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/v2/settings/scorecards",
		"/v2/settings/trackers",
		"/v2/stats/activity/scorecards",
		"/v2/data-privacy/data-for-email-address",
		"/v2/data-privacy/data-for-phone-number",
		"/v2/crm/entities",
		"/v2/crm/entity-schema",
		"/v2/crm/request-status",
		"/v2/library/folder-content",
		"/v2/calls/manual-crm-associations",
		"/v2/crm/integrations",
		"/v2/permission-profile", // GET, single item
		"/v2/permission-profile/users",
		"/v2/users/extensive",
		"/v2/library/folders",
		//
		// Endpoints that require query parameters.
		//
		// Requires query parameter: workspaceId.
		// Response field is `profiles`.
		"/v2/all-permission-profiles",
		// Requires query parameters: workspace-id, manager-id, from, to.
		"/v2/coaching",
		// Requires query parameters: logType, fromDateTime, toDateTime, cursor.
		"/v2/logs",
		// Requires query parameters: flowOwnerEmail.
		"/v2/flows",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithParameterFilterGetMethod(func(objectName string, operation *openapi3.Operation) bool {
			return queryParamFilterExceptions.Has(objectName) ||
				api3.OnlyOptionalQueryParameters(objectName, operation)
		}),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil, api3.IdenticalObjectLocator,
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
			schemas.Add("", object.ObjectName, object.DisplayName, field, object.URLPath, object.ResponseKey, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
