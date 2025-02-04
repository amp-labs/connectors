package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/asana/metadata"
	"github.com/amp-labs/connectors/providers/asana/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var ignoreEndpoints = []string{ // nolint:gochecknoglobals
	"/allocations",
	"/attachments",
	"/custom_fields",
	"/custom_field_settings",
	"/goals",
	"/goal_relationships",
	"/memberships",
	"/organization_export_gid",
	"/portfolios",
	"/portfolio_memberships",
	"/tasks",
	"/project_briefs",
	"/project_memberships",
	"/project_statuses",
	"/project_templates",
	"/project_statuses",
	"/sections",
	"/status_updates",
	"/stories",
	"/tasks",
	"/task_templates",
	"/teams",
	"/team_memberships",
	"/time_periods",
	"/task_gid",
	"/workspace_gid",
	"/user_task_lists",
	"/webhooks",
	"/workspace_memberships",
	"/events",
}

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CapitalizeFirstLetterEveryWord,
		),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil, api3.DataObjectLocator,
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range readObjects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field: field,
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
