// Extracts list endpoint schemas from the BambooHR OpenAPI spec and writes
// providers/bamboohr/internal/metadata/schemas.json.
package main

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/bamboohr/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

//nolint:gochecknoglobals
var outputMetadata = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
	fileconv.NewPath("providers/bamboohr/internal/metadata"),
)

type manualObject struct {
	ObjectName  string                          `json:"objectName"`
	DisplayName string                          `json:"displayName"`
	Path        string                          `json:"path"`
	ResponseKey string                          `json:"responseKey"`
	Fields      staticschema.FieldMetadataMapV2 `json:"fields"`
}

// Supported GET collection endpoints for the deep connector (expand over time).
//
//nolint:gochecknoglobals
var supportedObjects = map[string]string{
	"/api/v1/employees/directory":               "employees/directory",
	"/api/v1/meta/users":                        "meta/users",
	"/api/v1/employees":                         "employees",
	"/api/v1/time-off/requests":                 "time-off/requests",
	"/api/v1/whos-out":                          "whos-out",
	"/api/v1/applicant_tracking/jobs":           "applicant_tracking/jobs",
	"/api/v1/applicant_tracking/applications":   "applicant_tracking/applications",
	"/api/v1/applicant_tracking/locations":      "applicant_tracking/locations",
	"/api/v1/applicant_tracking/statuses":       "applicant_tracking/statuses",
	"/api/v1/applicant_tracking/hiring_leads":   "applicant_tracking/hiring_leads",
	"/api/v1/meta/fields":                       "meta/fields",
	"/api/v1/meta/lists":                        "meta/lists",
	"/api/v1/meta/tables":                       "meta/tables",
	"/api/v1/meta/time_off/policies":            "meta/time_off/policies",
	"/api/v1/company_information":               "company_information",
	"/api/v1/webhooks":                          "webhooks",
	"/api/v1/time-tracking/projects":            "time-tracking/projects",
	"/api/v1/hris/custom-fields":                "hris/custom-fields",
	"/api/v1/hris/custom-tables":                "hris/custom-tables",
	"/api/v1/hris/org/locations":                "hris/org/locations",
	"/api/v1/calendar-events":                   "calendar-events",
	"/api/v1/new-hire-packets":                  "new-hire-packets",
	"/api/v1/holidays":                          "holidays",
	"/api/v1/scheduling/schedules":              "scheduling/schedules",
	"/api/v1/scheduling/shifts":                 "scheduling/shifts",
	"/api/v1/time-tracking/break-policies":      "time-tracking/break-policies",
	"/api/v1/time-tracking/clock-entries":       "time-tracking/clock-entries",
	"/api/v1/time-tracking/configurations":      "time-tracking/configurations",
	"/api/v1/time-tracking/hour-entries":        "time-tracking/hour-entries",
	"/api/v1/time-tracking/shift-differentials": "time-tracking/shift-differentials",
	"/api/v1/training/category":                 "training/category",
	"/api/v1/training/type":                     "training/type",
	"/api/v1/alert-configurations":              "alert-configurations",
	"/api/v1/compensation/planning_cycles":      "compensation/planning_cycles",
}

//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	"employees/directory":               "Employees Directory",
	"meta/users":                        "Users",
	"employees":                         "Employees",
	"time-off/requests":                 "Time Off Requests",
	"whos-out":                          "Who's Out",
	"applicant_tracking/jobs":           "Job Openings",
	"applicant_tracking/applications":   "Applications",
	"applicant_tracking/locations":      "Locations",
	"applicant_tracking/statuses":       "Application Statuses",
	"applicant_tracking/hiring_leads":   "Hiring Leads",
	"meta/fields":                       "Fields",
	"meta/lists":                        "Lists",
	"meta/tables":                       "Tables",
	"meta/time_off/policies":            "Time Off Policies",
	"company_information":               "Company Information",
	"time-tracking/projects":            "Time Tracking Projects",
	"hris/custom-fields":                "Custom Fields",
	"hris/custom-tables":                "Custom Tables",
	"hris/org/locations":                "Org Locations",
	"calendar-events":                   "Calendar Events",
	"new-hire-packets":                  "New Hire Packets",
	"holidays":                          "Holidays",
	"scheduling/schedules":              "Schedules",
	"scheduling/shifts":                 "Shifts",
	"time-tracking/break-policies":      "Break Policies",
	"time-tracking/clock-entries":       "Clock Entries",
	"time-tracking/configurations":      "Time Tracking Configurations",
	"time-tracking/hour-entries":        "Hour Entries",
	"time-tracking/shift-differentials": "Shift Differentials",
	"training/category":                 "Training Categories",
	"training/type":                     "Training Types",
	"alert-configurations":              "Alert Configurations",
	"compensation/planning_cycles":      "Compensation Planning Cycles",
}

//nolint:gochecknoglobals
var objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
	"employees":                         "data",
	"employees/directory":               "employees",
	"time-off/requests":                 "data",
	"whos-out":                          "data",
	"holidays":                          "data",
	"scheduling/schedules":              "data",
	"scheduling/shifts":                 "data",
	"time-tracking/break-policies":      "data",
	"time-tracking/clock-entries":       "data",
	"time-tracking/configurations":      "data",
	"time-tracking/hour-entries":        "data",
	"time-tracking/shift-differentials": "data",
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

	explorer, err := files.FileManager.GetExplorer(
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

	goutils.MustBeNil(outputMetadata.FlushSchemas(schemas))
	goutils.MustBeNil(outputMetadata.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(supportedObjects))
}

func addManualObjects(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]) {
	var objects []manualObject

	err := json.Unmarshal(files.ManualObjects, &objects)
	goutils.MustBeNil(err)

	for _, object := range objects {
		schemas.Add("", object.ObjectName, object.DisplayName, object.Path, object.ResponseKey,
			object.Fields, nil, nil)
	}
}
