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
	objectNameH322Devices   = "h323_devices"    // nolint: gochecknoglobals
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = common.ModuleObjectNameToFieldName{ // nolint: gochecknoglobals

	ModuleMeeting: datautils.NewDefaultMap(map[string]string{
		"device_groups":     "groups",
		"archive_files":     "meetings",
		"meeting_summaries": "summaries",
		"billing_report":    "billing_reports",
		"activities_report": "activity_logs",
		"h323_devices":      "devices",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	),
	ModuleUser: datautils.NewDefaultMap(map[string]string{
		"contacts_groups": "groups",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	),
}

var supportedObjectsByWrite = map[common.ModuleID]datautils.StringSet{ // nolint: gochecknoglobals
	ModuleUser: datautils.NewSet(
		ObjectNameContactGroup,
		ObjectNameUser,
		ObjectNameGroup,
	),

	ModuleMeeting: datautils.NewSet(
		objectNameTrackingField,
		objectNameDevice,
		objectNameH322Devices,
	),
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	ObjectNameContactGroup:  "/contacts/groups",
	ObjectNameUser:          "/users",
	ObjectNameGroup:         "/groups",
	objectNameTrackingField: "/tracking_fields",
	objectNameDevice:        "/devices",
	objectNameH322Devices:   "/h323/devices",
}, func(objectName string) (path string) {
	return objectName
},
)

// Each object has different fields that represent the record id.
// This map is used to get the record id field for each object.
var objectNameToWriteResponseIdentifier = common.ModuleObjectNameToFieldName{ // nolint: gochecknoglobals

	ModuleMeeting: datautils.NewDefaultMap(map[string]string{
		objectNameTrackingField: "id",
		objectNameDevice:        "",
		objectNameH322Devices:   "id",
	},
		func(objectName string) (fieldName string) {
			return "id"
		},
	),

	ModuleUser: datautils.NewDefaultMap(map[string]string{
		ObjectNameContactGroup: "group_id",
		ObjectNameUser:         "id",
		ObjectNameGroup:        "",
	},
		func(objectName string) (fieldName string) {
			return "id"
		},
	),
}
