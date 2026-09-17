// Extracts list endpoint schemas from the BambooHR OpenAPI spec and writes providers/bambooHR/metadata/schemas.json.
package main

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/bambooHR/metadata"
	"github.com/amp-labs/connectors/providers/bambooHR/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

//nolint:gochecknoglobals
var (
	//go:embed manual-objects.json
	manualObjectsJSON string
)

type manualObject struct {
	ObjectName  string                           `json:"objectName"`
	DisplayName string                           `json:"displayName"`
	Path        string                           `json:"path"`
	ResponseKey string                           `json:"responseKey"`
	Fields      staticschema.FieldMetadataMapV2 `json:"fields"`
}

// Supported GET collection endpoints for the deep connector (expand over time).
//
//nolint:gochecknoglobals
var supportedObjects = map[string]string{
	"/api/v1/employees/directory":             "employees_directory",
	"/api/v1/meta/users":                      "meta_users",
	"/api/v1/employees":                       "employees",
	"/api/v1/time-off/requests":               "time_off_requests",
	"/api/v1/whos-out":                        "whos_out",
	"/api/v1/applicant_tracking/jobs":         "applicant_tracking_jobs",
	"/api/v1/applicant_tracking/applications": "applicant_tracking_applications",
	"/api/v1/applicant_tracking/locations":    "applicant_tracking_locations",
	"/api/v1/applicant_tracking/statuses":     "applicant_tracking_statuses",
	"/api/v1/applicant_tracking/hiring_leads": "applicant_tracking_hiring_leads",
	"/api/v1/meta/fields":                     "meta_fields",
	"/api/v1/meta/lists":                      "meta_lists",
	"/api/v1/meta/tables":                     "meta_tables",
	"/api/v1/meta/time_off/policies":          "meta_time_off_policies",
	"/api/v1/company_information":             "company_information",
	"/api/v1/webhooks":                        "webhooks",
	"/api/v1/time-tracking/projects":          "time_tracking_projects",
	"/api/v1/hris/custom-fields":              "hris_custom_fields",
	"/api/v1/hris/custom-tables":              "hris_custom_tables",
	"/api/v1/hris/org/locations":              "hris_org_locations",
	"/api/v1/calendar-events":                 "calendar_events",
	"/api/v1/new-hire-packets":                "new_hire_packets",
}

//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	"employees_directory":             "Employees Directory",
	"meta_users":                      "Users",
	"employees":                       "Employees",
	"time_off_requests":               "Time Off Requests",
	"whos_out":                        "Who's Out",
	"applicant_tracking_jobs":         "Job Openings",
	"applicant_tracking_applications": "Applications",
	"applicant_tracking_locations":    "Locations",
	"applicant_tracking_statuses":     "Application Statuses",
	"applicant_tracking_hiring_leads": "Hiring Leads",
	"meta_fields":                     "Fields",
	"meta_lists":                      "Lists",
	"meta_tables":                     "Tables",
	"meta_time_off_policies":          "Time Off Policies",
	"company_information":             "Company Information",
	"time_tracking_projects":          "Time Tracking Projects",
	"hris_custom_fields":              "Custom Fields",
	"hris_custom_tables":              "Custom Tables",
	"hris_org_locations":              "Org Locations",
	"calendar_events":                 "Calendar Events",
	"new_hire_packets":                "New Hire Packets",
}

//nolint:gochecknoglobals
var objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
	"employees":           "data",
	"employees_directory": "employees",
	"time_off_requests":   "timeOffRequests",
	"whos_out":            "employees",
},
	func(objectName string) string {
		return ""
	},
)

func pathToConnectorPath(path string) string {
	trimmed, ok := strings.CutPrefix(path, "/api/v1")
	if !ok {
		return path
	}

	if trimmed == "" {
		return "/"
	}

	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}

	return trimmed
}

func main() {
	allowPaths := make([]string, 0, len(supportedObjects))
	for path := range supportedObjects {
		allowPaths = append(allowPaths, path)
	}

	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			func(displayName string) string {
				return strings.ReplaceAll(displayName, "_", " ")
			},
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			api3.Pluralize,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowPaths),
		supportedObjects,
		displayNameOverrides,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		urlPath := pathToConnectorPath(object.URLPath)

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, urlPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	addManualObjects(schemas)

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(supportedObjects))
}

func addManualObjects(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]) {
	var objects []manualObject

	err := json.Unmarshal([]byte(manualObjectsJSON), &objects)
	goutils.MustBeNil(err)

	for _, object := range objects {
		schemas.Add("", object.ObjectName, object.DisplayName, object.Path, object.ResponseKey,
			object.Fields, nil, nil)
	}
}
