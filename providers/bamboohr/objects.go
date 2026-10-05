package bamboohr

import "github.com/amp-labs/connectors/internal/datautils"

// Object names for read routing (match schemas.json keys).
const (
	objectApplicantTrackingApplications  = "applicant_tracking/applications"
	objectApplicantTrackingHiringLeads   = "applicant_tracking/hiring_leads"
	objectApplicantTrackingJobs          = "applicant_tracking/jobs"
	objectApplicantTrackingLocations     = "applicant_tracking/locations"
	objectApplicantTrackingStatuses      = "applicant_tracking/statuses"
	objectCalendarEvents                 = "calendar-events"
	objectEmployees                      = "employees"
	objectEmployeesDirectory             = "employees/directory"
	objectHRISCustomFields               = "hris/custom-fields"
	objectHRISCustomTables               = "hris/custom-tables"
	objectHRISOrgLocations               = "hris/org/locations"
	objectMetaFields                     = "meta/fields"
	objectMetaLists                      = "meta/lists"
	objectMetaTables                     = "meta/tables"
	objectMetaTimeOffPolicies            = "meta/time_off/policies"
	objectMetaUsers                      = "meta/users"
	objectNewHirePackets                 = "new-hire-packets"
	objectTimeOffRequests                = "time-off/requests"
	objectTimeTrackingProjects           = "time-tracking/projects"
	objectWebhooks                       = "webhooks"
	objectWhosOut                        = "whos-out"
	objectHolidays                       = "holidays"
	objectSchedulingSchedules            = "scheduling/schedules"
	objectSchedulingShifts               = "scheduling/shifts"
	objectTimeTrackingBreakPolicies      = "time-tracking/break-policies"
	objectTimeTrackingClockEntries       = "time-tracking/clock-entries"
	objectTimeTrackingConfigurations     = "time-tracking/configurations"
	objectTimeTrackingHourEntries        = "time-tracking/hour-entries"
	objectTimeTrackingShiftDifferentials = "time-tracking/shift-differentials"
	objectTrainingCategory               = "training/category"
	objectTrainingType                   = "training/type"
	objectAlertConfigurations            = "alert-configurations"
	objectCompensationPlanningCycles     = "compensation/planning_cycles"
)

//nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectApplicantTrackingApplications,
	objectApplicantTrackingHiringLeads,
	objectApplicantTrackingJobs,
	objectApplicantTrackingLocations,
	objectApplicantTrackingStatuses,
	objectCalendarEvents,
	objectEmployees,
	objectEmployeesDirectory,
	objectHRISCustomFields,
	objectHRISCustomTables,
	objectHRISOrgLocations,
	objectMetaFields,
	objectMetaLists,
	objectMetaTables,
	objectMetaTimeOffPolicies,
	objectMetaUsers,
	objectNewHirePackets,
	objectTimeOffRequests,
	objectTimeTrackingProjects,
	objectWebhooks,
	objectWhosOut,
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
	objectAlertConfigurations,
)

//nolint:gochecknoglobals
var pageNumberReadObjects = datautils.NewStringSet(
	objectApplicantTrackingApplications,
	objectCalendarEvents,
	objectHRISCustomFields,
	objectHRISCustomTables,
	objectHRISOrgLocations,
	objectNewHirePackets,
	objectTimeOffRequests,
	objectTimeTrackingProjects,
	objectWhosOut,
	objectHolidays,
	objectSchedulingSchedules,
	objectSchedulingShifts,
	objectTimeTrackingClockEntries,
	objectTimeTrackingConfigurations,
	objectTimeTrackingHourEntries,
	objectTimeTrackingShiftDifferentials,
)

//nolint:gochecknoglobals
var pageSizeReadObjects = datautils.NewStringSet(
	objectCalendarEvents,
	objectHRISCustomFields,
	objectHRISCustomTables,
	objectHRISOrgLocations,
	objectNewHirePackets,
	objectTimeOffRequests,
	objectTimeTrackingProjects,
	objectWhosOut,
	objectHolidays,
	objectSchedulingSchedules,
	objectSchedulingShifts,
	objectTimeTrackingClockEntries,
	objectTimeTrackingConfigurations,
	objectTimeTrackingHourEntries,
	objectTimeTrackingShiftDifferentials,
)

//nolint:gochecknoglobals
var cursorReadObjects = datautils.NewStringSet(
	objectEmployees,
)
