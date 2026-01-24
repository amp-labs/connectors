package meeting

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/zoom/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	allowedEndpoints = []string{ // nolint:gochecknoglobals
		"/archive_files",
		"/devices/groups",
		"/devices",
		"/meetings/meeting_summaries",
		"/report/billing",
		"/report/activities",
		"/sip_phones/phones",
		"/tracking_fields",
		"/h323/devices",
		"/users/{userId}/recordings",
		"/users/{userId}/meetings",
		"/users/{userId}/upcoming_meetings",
		"/report/users/{userId}/meetings",
		"/users/{userId}/meeting_templates",
		"/users/{userId}/tsp",
		"/users/{userId}/webinars",
		"/report/users",
		"/report/cloud_recording",
		"/report/operationlogs",
		"/report/meeting_activities",
		"/report/telephone",
		"/report/upcoming_events",
	}

	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/devices/groups":                 "device_groups",
		"/archive_files":                  "archive_files",
		"/meetings/meeting_summaries":     "meeting_summaries",
		"/report/billing":                 "billing",
		"/report/activities":              "activities",
		"/h323/devices":                   "h323_devices",
		"/report/users/{userId}/meetings": "meetings_report",
		"/report/users":                   "users_report",
		"/report/cloud_recording":         "recordings_report",
		"/report/operationlogs":           "operation_logs_report",
		"/report/meeting_activities":      "meeting_activities_report",
		"/report/telephone":               "telephone_report",
		"/report/upcoming_events":         "upcoming_events_report",
	}

	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"devices/groups":                  "Device Groups",
		"h323/devices":                    "H.323/SIP Devices",
		"/report/users/{userId}/meetings": "Meetings Report",
		"/report/users":                   "Users Report",
		"/report/cloud_recording":         "Recordings Report",
		"/report/operationlogs":           "Operation Logs Report",
		"/report/meeting_activities":      "Meeting Activities Report",
		"/report/telephone":               "Telephone Report",
		"/report/upcoming_events":         "Upcoming Events Report",
	}
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.MeetingFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)

	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(
		"GET",
		api3.NewAllowPathStrategy(allowedEndpoints),
		objectEndpoints, displayNameOverride,
		nil,
	)

	goutils.MustBeNil(err)

	return objects
}
