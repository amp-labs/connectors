package zoom

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

var incrementalObjects = datautils.NewSet( //nolint:gochecknoglobals
	"activities",
	"archive_files",
	"meeting_activities_report",
	"meeting_summaries",
	"operation_logs_report",
	"recordings_report",
	"telephone_report",
	"upcoming_events_report",
	"users_report",
	// User scoped objects.
	core.ObjectNameMeetings.All,
	core.ObjectNameMeetings.Self,
	core.ObjectNameMeetingsReport.All,
	core.ObjectNameMeetingsReport.Self,
	core.ObjectNameRecordings.All,
	core.ObjectNameRecordings.Self,
)

// mandatoryDateObjects defines which objects require mandatory from/to query parameters.
// These endpoints will get default 29-day range when Since/Until are not provided.
var mandatoryDateObjects = datautils.NewSet( //nolint:gochecknoglobals
	"meeting_activities_report",
	"operation_logs_report",
	"recordings_report",
	"telephone_report",
	"upcoming_events_report",
	"users_report",
	// User scoped objects.
	core.ObjectNameMeetingsReport.All,
	core.ObjectNameMeetingsReport.Self,
	core.ObjectNameRecordings.All,
	core.ObjectNameRecordings.Self,
)

var objectNameToResponseField = map[string]string{ // nolint: gochecknoglobals
	"activities_report":         "activity_logs",
	"archive_files":             "meetings",
	"billing_report":            "billing_reports",
	"contacts_groups":           "groups",
	"devices":                   "devices",
	"devices_groups":            "groups",
	"groups":                    "groups",
	"h323_devices":              "devices",
	"meeting_activities_report": "meeting_activity_logs",
	"meeting_summaries":         "summaries",
	"operation_logs_report":     "operation_logs",
	"phones":                    "phones",
	"recordings_report":         "cloud_recording_storage",
	"telephone_report":          "telephony_usage",
	"tracking_fields":           "tracking_fields",
	"upcoming_events_report":    "upcoming_events",
	"upcoming_meetings":         "meetings",
	"users":                     "users",
	"users_report":              "users",
	"webinars":                  "webinars",
	// User scoped objects.
	core.ObjectNameMeetingTemplates.All:  "templates",
	core.ObjectNameMeetingTemplates.Self: "templates",
	core.ObjectNameMeetings.All:          "meetings",
	core.ObjectNameMeetings.Self:         "meetings",
	core.ObjectNameMeetingsReport.All:    "meetings",
	core.ObjectNameMeetingsReport.Self:   "meetings",
	core.ObjectNameRecordings.All:        "meetings",
	core.ObjectNameRecordings.Self:       "meetings",
	core.ObjectNameTsp.All:               "tsp_accounts",
	core.ObjectNameTsp.Self:              "tsp_accounts",
}

var objectNameToReadResponseIdentifier = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	"activities_report":         "ip_address",
	"contacts_groups":           "group_id",
	"devices":                   "device_id",
	"devices_groups":            "name",
	"meeting_activities_report": "activity_time",
	"meeting_summaries":         "meeting_id",
	"operation_logs_report":     "time",
	"phones":                    "phone_id",
	"recordings_report":         "date",
	"telephone_report":          "phone_number",
}, func(objectName string) string {
	return "id"
})

// Each object has different fields that represent the record id.
// This map is used to get the record id field for each object.
var objectNameToWriteResponseIdentifier = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	core.ObjectNameContactGroup:  "group_id",
	core.ObjectNameDevice:        "",
	core.ObjectNameGroup:         "",
	core.ObjectNameH322Device:    "id",
	core.ObjectNameTrackingField: "id",
	core.ObjectNameUser:          "id",
},
	func(objectName string) (fieldName string) {
		return "id"
	},
)
