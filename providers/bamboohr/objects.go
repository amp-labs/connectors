package bamboohr

import "github.com/amp-labs/connectors/internal/datautils"

// Object names for read routing (match schemas.json keys).
const (
	objectApplicantTrackingApplications  = "applicant_tracking_applications"
	objectApplicantTrackingHiringLeads   = "applicant_tracking_hiring_leads"
	objectApplicantTrackingJobs          = "applicant_tracking_jobs"
	objectApplicantTrackingLocations     = "applicant_tracking_locations"
	objectApplicantTrackingStatuses      = "applicant_tracking_statuses"
	objectCalendarEvents                 = "calendar_events"
	objectEmployees                      = "employees"
	objectEmployeesDirectory             = "employees_directory"
	objectHRISCustomFields               = "hris_custom_fields"
	objectHRISCustomTables               = "hris_custom_tables"
	objectHRISOrgLocations               = "hris_org_locations"
	objectMetaFields                     = "meta_fields"
	objectMetaLists                      = "meta_lists"
	objectMetaTables                     = "meta_tables"
	objectMetaTimeOffPolicies            = "meta_time_off_policies"
	objectMetaUsers                      = "meta_users"
	objectNewHirePackets                 = "new_hire_packets"
	objectTimeOffRequests                = "time_off_requests"
	objectTimeTrackingProjects           = "time_tracking_projects"
	objectWebhooks                       = "webhooks"
	objectWhosOut                        = "whos_out"
	objectHolidays                       = "holidays"
	objectSchedulingSchedules            = "scheduling_schedules"
	objectSchedulingShifts               = "scheduling_shifts"
	objectTimeTrackingBreakPolicies      = "time_tracking_break_policies"
	objectTimeTrackingClockEntries       = "time_tracking_clock_entries"
	objectTimeTrackingConfigurations     = "time_tracking_configurations"
	objectTimeTrackingHourEntries        = "time_tracking_hour_entries"
	objectTimeTrackingShiftDifferentials = "time_tracking_shift_differentials"
	objectTrainingCategory               = "training_category"
	objectTrainingType                   = "training_type"
	objectAlertConfigurations            = "alert_configurations"
	objectCompensationPlanningCycles     = "compensation_planning_cycles"
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
