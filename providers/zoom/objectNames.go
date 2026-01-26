package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

var supportedObjectsByRead = metadata.Schemas.ObjectNames() // nolint: gochecknoglobals

var (
	ObjectNameContactGroup  = "contacts_groups" // nolint: gochecknoglobals
	ObjectNameUser          = "users"           // nolint: gochecknoglobals
	ObjectNameGroup         = "groups"          // nolint: gochecknoglobals
	objectNameTrackingField = "tracking_fields" // nolint: gochecknoglobals
	objectNameDevice        = "devices"         // nolint: gochecknoglobals
	objectNameH322Device    = "h323_devices"    // nolint: gochecknoglobals
	objectNameMeeting       = "meetings"        // nolint: gochecknoglobals
	objectNameTsp           = "tsp"             // nolint: gochecknoglobals
	objectNameWebinar       = "webinars"        // nolint: gochecknoglobals
)

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
// These endpoints will get default 30-day range when Since/Until are not provided.
var mandatoryDateObjects = datautils.NewSet( //nolint:gochecknoglobals
	"users_report",
	"recordings_report",
	"meetings_report",
	"operation_logs_report",
	"meeting_activities_report",
	"telephone_report",
	"upcoming_events_report",
)

var supportedObjectsByWrite = map[common.ModuleID]datautils.StringSet{ // nolint: gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		ObjectNameContactGroup,
		ObjectNameUser,
		ObjectNameGroup,
		objectNameTrackingField,
		objectNameDevice,
		objectNameH322Device,
		objectNameTsp,
		objectNameWebinar,
	),
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	ObjectNameContactGroup:  "/contacts/groups",
	ObjectNameUser:          "/users",
	ObjectNameGroup:         "/groups",
	objectNameTrackingField: "/tracking_fields",
	objectNameDevice:        "/devices",
	objectNameH322Device:    "/h323/devices",
	objectNameMeeting:       "/users/me/meetings",
	objectNameTsp:           "/users/me/tsp",
	objectNameWebinar:       "/users/me/webinars",
}, func(objectName string) (path string) {
	return objectName
},
)

// Each object has different fields that represent the record id.
// This map is used to get the record id field for each object.
var objectNameToWriteResponseIdentifier = common.ModuleObjectNameToFieldName{ // nolint: gochecknoglobals
	common.ModuleRoot: datautils.NewDefaultMap(map[string]string{
		objectNameTrackingField: "id",
		objectNameDevice:        "",
		objectNameH322Device:    "id",
		ObjectNameContactGroup:  "group_id",
		ObjectNameUser:          "id",
		ObjectNameGroup:         "",
	},
		func(objectName string) (fieldName string) {
			return "id"
		},
	),
}
