package bamboohr

import "github.com/amp-labs/connectors/internal/datautils"

// Object names for read routing (match schemas.json keys).
const (
	objectApplicantTrackingApplications = "applicant_tracking_applications"
	objectApplicantTrackingHiringLeads  = "applicant_tracking_hiring_leads"
	objectApplicantTrackingJobs         = "applicant_tracking_jobs"
	objectApplicantTrackingLocations    = "applicant_tracking_locations"
	objectApplicantTrackingStatuses     = "applicant_tracking_statuses"
	objectCalendarEvents                = "calendar_events"
	objectEmployees                     = "employees"
	objectEmployeesDirectory            = "employees_directory"
	objectHRISCustomFields              = "hris_custom_fields"
	objectHRISCustomTables              = "hris_custom_tables"
	objectHRISOrgLocations              = "hris_org_locations"
	objectMetaFields                    = "meta_fields"
	objectMetaLists                     = "meta_lists"
	objectMetaTables                    = "meta_tables"
	objectMetaTimeOffPolicies           = "meta_time_off_policies"
	objectMetaUsers                     = "meta_users"
	objectNewHirePackets                = "new_hire_packets"
	objectTimeOffRequests               = "time_off_requests"
	objectTimeTrackingProjects          = "time_tracking_projects"
	objectWebhooks                      = "webhooks"
	objectWhosOut                       = "whos_out"
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
)

//nolint:gochecknoglobals
var cursorReadObjects = datautils.NewStringSet(
	objectEmployees,
)
