package zoom

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
	"github.com/amp-labs/connectors/providers/zoom/internal/metadata"
)

var supportedObjectsByRead = metadata.Catalog.Objects.KeySet() // nolint: gochecknoglobals

var incrementalObjects = datautils.NewSet( //nolint:gochecknoglobals
	"recordings",
	"archive_files",
	"meeting_summaries",
	"activities",
	"meetings",
	"users_report",
	"recordings_report",
	"meetings_report",
	"operation_logs_report",
	"meeting_activities_report",
	"telephone_report",
	"upcoming_events_report",
)

// mandatoryDateObjects defines which objects require mandatory from/to query parameters.
// These endpoints will get default 29-day range when Since/Until are not provided.
var mandatoryDateObjects = datautils.NewSet( //nolint:gochecknoglobals
	"users_report",
	"recordings_report",
	"meetings_report",
	"operation_logs_report",
	"meeting_activities_report",
	"telephone_report",
	"upcoming_events_report",
	"recordings",
)

var supportedObjectsByWrite = datautils.NewSet( // nolint: gochecknoglobals
	core.ObjectNameContactGroup,
	core.ObjectNameUser,
	core.ObjectNameGroup,
	core.ObjectNameTrackingField,
	core.ObjectNameDevice,
	core.ObjectNameH322Device,
	core.ObjectNameTsp,
	core.ObjectNameWebinar,
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
	"meeting_templates":         "templates",
	"meetings":                  "meetings",
	"meetings_report":           "meetings",
	"operation_logs_report":     "operation_logs",
	"phones":                    "phones",
	"recordings":                "meetings",
	"recordings_report":         "cloud_recording_storage",
	"telephone_report":          "telephony_usage",
	"tracking_fields":           "tracking_fields",
	"tsp":                       "tsp_accounts",
	"upcoming_events_report":    "upcoming_events",
	"upcoming_meetings":         "meetings",
	"users":                     "users",
	"users_report":              "users",
	"webinars":                  "webinars",
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
	core.ObjectNameTrackingField: "id",
	core.ObjectNameDevice:        "",
	core.ObjectNameH322Device:    "id",
	core.ObjectNameContactGroup:  "group_id",
	core.ObjectNameUser:          "id",
	core.ObjectNameGroup:         "",
},
	func(objectName string) (fieldName string) {
		return "id"
	},
)
