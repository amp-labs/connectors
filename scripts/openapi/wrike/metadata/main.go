// Generates providers/wrike/metadata/schemas.json from the Wrike API v4
// OpenAPI spec (scripts/openapi/wrike/internal/files/wrike_api_v4_ver154.json).
//
// Run from the repo root:
//
//	go run ./scripts/openapi/wrike/metadata
//
// Objects are curated explicitly (allow-list): account-level list endpoints
// whose response is the standard Wrike envelope {kind, data: [...]}.
//
// Excluded on purpose (see providers/wrike/README.md):
//   - Singletons and lookups: /account, /version, /colors, /ids,
//     /whiteboard_ids, /data_export, /data_export_schema, /async_job.
//   - Lists that require query parameters to be called at all:
//     /user_schedule_capacity_change, /user_schedule_partial_exclusion
//     (userIds + dateRange).
//   - Nested lists needing a parent id (/folders/{id}/tasks, .../rollups,
//     .../dependencies, .../hourly_rates, .../timelog_lock_periods, *_history).
package main

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/wrike/metadata"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/scripts/openapi/wrike/internal/files"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// The spec's server URL is https://www.wrike.com/api/v4 while the catalog base
// URL is https://www.wrike.com/api, so the version is prepended to every path.
const apiVersionPrefix = "/v4"

// supportedObjects maps spec URL paths to object names. Names are the last URL
// segment, which is also how Wrike's documentation refers to each resource.
//
//nolint:gochecknoglobals
var supportedObjects = map[string]string{
	"/access_roles":               "access_roles",
	"/approvals":                  "approvals",
	"/attachments":                "attachments",
	"/audit_log":                  "audit_log",
	"/bookings":                   "bookings",
	"/comments":                   "comments",
	"/contacts":                   "contacts",
	"/custom_item_types":          "custom_item_types",
	"/customfields":               "customfields",
	"/folder_blueprints":          "folder_blueprints",
	"/folders":                    "folders",
	"/groups":                     "groups",
	"/invitations":                "invitations",
	"/jobroles":                   "jobroles",
	"/placeholders":               "placeholders",
	"/request_forms":              "request_forms",
	"/spaces":                     "spaces",
	"/task_blueprints":            "task_blueprints",
	"/tasks":                      "tasks",
	"/timelog_categories":         "timelog_categories",
	"/timelogs":                   "timelogs",
	"/timesheet_submission_rules": "timesheet_submission_rules",
	"/timesheets":                 "timesheets",
	"/user_schedule_exclusions":   "user_schedule_exclusions",
	"/user_types":                 "user_types",
	"/workflows":                  "workflows",
	"/workschedules":              "workschedules",
}

// Names the display-name processors cannot derive from the path.
//
//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	"customfields":             "Custom Fields",
	"folders":                  "Folders & Projects",
	"jobroles":                 "Job Roles",
	"timelogs":                 "Timelogs",
	"user_schedule_exclusions": "User Schedule Exceptions",
	"workschedules":            "Work Schedules",
}

func pathsOf(m map[string]string) []string {
	paths := make([]string, 0, len(m))
	for path := range m {
		paths = append(paths, path)
	}

	return paths
}

func main() {
	explorer, err := files.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			func(displayName string) string {
				return strings.ReplaceAll(displayName, "_", " ")
			},
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	// Every list response is the Wrike envelope {kind, data: [...]}.
	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(pathsOf(supportedObjects)),
		supportedObjects,
		displayNameOverrides,
		api3.DataObjectLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"urlPath", object.URLPath,
				"error", object.Problem,
			)

			continue
		}

		urlPath := apiVersionPrefix + object.URLPath

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName,
				urlPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	addLiveFields(schemas)

	// FlushSchemas keeps full paths; SaveSchemas would factor the common "/v4/"
	// prefix out into the deprecated module path.
	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(objects), "expected", len(supportedObjects))
}

type liveFields struct {
	ObjectName string                          `json:"objectName"`
	Fields     staticschema.FieldMetadataMapV2 `json:"fields"`
}

// addLiveFields merges undocumented-but-observed fields (files.LiveFields) into existing objects.
func addLiveFields(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]) {
	var overlays []liveFields

	goutils.MustBeNil(json.Unmarshal(files.LiveFields, &overlays))

	objects := schemas.Modules[common.ModuleRoot].Objects

	for _, overlay := range overlays {
		object, ok := objects[overlay.ObjectName]
		if !ok {
			slog.Error("live field overlay targets unknown object", "objectName", overlay.ObjectName)

			continue
		}

		schemas.Add(common.ModuleRoot, overlay.ObjectName, object.DisplayName,
			object.URLPath, object.ResponseKey, overlay.Fields, nil, nil)
	}
}
