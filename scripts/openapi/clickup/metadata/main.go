package main

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/clickup/internal/files"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/merging"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
)

// nolint:gochecknoglobals
var (
	ignoreV2 = filters.NewRules(
		// OpenAPI is malformed and doesn't have schema definition for views.
		// The object is created manually (see manual-objects.json) using the shape of example response.
		"/v2/team/{team_id}/view",
		// Endpoints that do not hold "Collections".
		"/v2/task/bulk_time_in_status/task_ids",
		"/v2/team/{team_id}/plan",
		"/v2/team/{team_id}/seats",
		"/v2/team/{team_id}/taskTemplate",
		"/v2/team/{team_id}/time_entries/current",
		"/v2/user",
		// Endpoints with multiple lists.
		"/v2/team/{team_id}/shared",
		// Ignore custom fields.
		"/v2/team/{team_id}/field",
		// Ignore webhooks.
		"/v2/team/{team_id}/webhook",
	)

	// Note: URLs have been normalized where all "team_id" ids are replaced with "workspace_id".
	displayNames = core.NewTrackedMap(map[string]string{
		"/v2/team":                            "Workspaces (Teams)",
		"/v2/team/{workspace_id}/custom_item": "Custom Task Types",
		"/v2/team/{workspace_id}/customroles": "Custom Roles",
		"/v3/workspaces/{workspace_id}/docs":  "Documents",
	})

	objectNameOverride = core.NewTrackedMap(map[string]string{
		"/v2/team/{workspace_id}/time_entries/tags": "time_entries/tags",
	})
)

func main() {
	schemasV2, schemasV3 := extractSchemas()

	pipeV2 := pipeline.NewSchemaPipe(schemasV2).
		Filter(filters.KeepWithPath(filters.AndPathMatcher{
			filters.NewDenyPathStrategy(ignoreV2),
			filters.OnlyAllowedIdsPath("team_id"),
		})).
		Map(replaceTeamIdWithWorkspaceId). // V3 calls "team id" a "workspace id". It is the same concept.
		Reduce(reducers.ShortestNameFromUrl)

	pipeV3 := pipeline.NewSchemaPipe(schemasV3).
		Filter(filters.KeepWithPath(filters.OnlyAllowedIdsPath("workspace_id"))).
		Map(mapping.RemoveUrlPrefix("/api")).
		Reduce(reducers.ShortestNameFromUrl)

	pipe := pipeline.Combine(pipeV2, pipeV3, merging.CombineByObjectName, merging.ChooseRight).
		Map(mapping.ObjectNameFormat(mapping.Pluralize)).
		Map(mapping.DisplayNameFromObjectName).
		Map(mapping.DisplayNameFormat(
			mapping.CamelCaseToSpaceSeparated,
			mapping.CapitalizeFirstLetterEveryWord,
		)).
		Map(mapping.DisplayNameOverride(displayNames)).
		Map(mapping.ObjectNameOverride(objectNameOverride))

	goutils.MustBeNil(output.WriteZippedMetadata(files.OutputDir, pipe, output.WithManualObjects(files.ManualObjects)))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
	output.PrintUnusedRules("ignoreV2", ignoreV2)
	output.PrintRedundantKeys("displayNames", displayNames)
	output.PrintRedundantKeys("objectNameOverride", objectNameOverride)
}

func extractSchemas() ([]spec.Schema, []spec.Schema) {
	extractorV2, err := files.OpenApiFileV2.Extractor()
	goutils.MustBeNil(err)
	extractorV3, err := files.OpenApiFileV3.Extractor()
	goutils.MustBeNil(err)

	schemasV2, err := extractorV2.ExtractListSchemas(api.GET,
		api.List.WithArrayLocator(func(objectName, fieldName string) bool {
			return datautils.NewDefaultMap[string, string](map[string]string{
				"/v2/group":                          "groups",
				"/v2/team":                           "teams",
				"/v2/team/{team_Id}/task":            "tasks",
				"/v2/team/{team_id}/custom_item":     "custom_items",
				"/v2/team/{team_id}/customroles":     "custom_roles",
				"/v2/team/{team_id}/folder_template": "templates",
				"/v2/team/{team_id}/goal":            "goals",
				"/v2/team/{team_id}/list_template":   "templates",
				"/v2/team/{team_id}/space":           "spaces",
			}, func(objectName string) string {
				return "data"
			}).Get(objectName) == fieldName
		}),
	)
	goutils.MustBeNil(err)
	schemasV3, err := extractorV3.ExtractListSchemas(api.GET,
		api.List.WithArrayLocator(func(objectName, fieldName string) bool {
			if objectName == "/api/v3/workspaces/{workspace_id}/docs" {
				return fieldName == "docs"
			}

			return fieldName == "data"
		}),
	)
	goutils.MustBeNil(err)

	return schemasV2, schemasV3
}

func replaceTeamIdWithWorkspaceId(schema spec.Schema) spec.Schema {
	url := schema.UrlPath
	url = strings.ReplaceAll(url, "{team_id}", "{workspace_id}")
	url = strings.ReplaceAll(url, "{team_Id}", "{workspace_id}")
	schema.UrlPath = url

	return schema
}
