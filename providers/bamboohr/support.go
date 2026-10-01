package bamboohr

import (
	"net/http"

	"github.com/amp-labs/connectors/internal/datautils"
)

// writeAndDeleteSupportedObjects support create, update, and HTTP DELETE on the collection item path.
// Paths and methods come from the BambooHR OpenAPI spec and match schemas.json keys.
//
//nolint:gochecknoglobals
var writeAndDeleteSupportedObjects = datautils.NewStringSet(
	objectWebhooks,
	objectTimeTrackingProjects,
	objectHRISCustomFields,
	objectHRISOrgLocations,
	objectNewHirePackets,
	objectHolidays,
	objectSchedulingSchedules,
	objectSchedulingShifts,
	objectTimeTrackingBreakPolicies,
	objectTimeTrackingClockEntries,
	objectTimeTrackingConfigurations,
	objectTimeTrackingHourEntries,
	objectTimeTrackingShiftDifferentials,
	objectTrainingCategory,
	objectTrainingType,
	objectCompensationPlanningCycles,
	objectEmployees,
)

// writeOnlySupportedObjects support create and update, but not generic DELETE.
// time_off_requests is cancelled via a separate action URL.
// hris_custom_tables and alert_configurations have no DELETE on the item path.
//
//nolint:gochecknoglobals
var writeOnlySupportedObjects = datautils.NewStringSet(
	objectTimeOffRequests,
	objectHRISCustomTables,
	objectAlertConfigurations,
)

// writeSupportedObjects is the union of write-capable objects.
//
//nolint:gochecknoglobals
var writeSupportedObjects = datautils.MergeSets(
	writeAndDeleteSupportedObjects,
	writeOnlySupportedObjects,
)

// deleteSupportedObjects lists objects with DELETE on {collection}/{recordId}.
// hris_custom_fields DELETE archives the field.
//
//nolint:gochecknoglobals
var deleteSupportedObjects = writeAndDeleteSupportedObjects

// writeUpdateMethod maps object name to the HTTP method used when RecordId is set.
// Unlisted objects use PUT. employees updates with POST on /employees/{id}.
//
//nolint:gochecknoglobals
var writeUpdateMethod = datautils.NewDefaultMap(map[string]string{
	objectWebhooks:                       http.MethodPut,
	objectTimeTrackingProjects:           http.MethodPatch,
	objectTimeOffRequests:                http.MethodPatch,
	objectHolidays:                       http.MethodPatch,
	objectSchedulingSchedules:            http.MethodPatch,
	objectSchedulingShifts:               http.MethodPatch,
	objectTimeTrackingBreakPolicies:      http.MethodPatch,
	objectTimeTrackingClockEntries:       http.MethodPatch,
	objectTimeTrackingConfigurations:     http.MethodPatch,
	objectTimeTrackingHourEntries:        http.MethodPatch,
	objectTimeTrackingShiftDifferentials: http.MethodPatch,
	objectEmployees:                      http.MethodPost,
}, func(string) string {
	return http.MethodPut
})

// writeUpdateContentType sets Content-Type for update requests when non-empty.
// time_off_requests expects JSON Merge Patch per OpenAPI.
//
//nolint:gochecknoglobals
var writeUpdateContentType = datautils.NewDefaultMap(map[string]string{
	objectTimeOffRequests: "application/merge-patch+json",
}, func(string) string {
	return "application/json"
})
