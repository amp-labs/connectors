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
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = common.ModuleObjectNameToFieldName{ // nolint: gochecknoglobals

	ModuleMeeting: datautils.NewDefaultMap(map[string]string{
		"device_groups":     "groups",
		"archive_files":     "meetings",
		"meeting_summaries": "summaries",
		"billing_report":    "billing_reports",
		"activities_report": "activity_logs",
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
	),
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	ObjectNameContactGroup:  "/contacts/groups",
	ObjectNameUser:          "/users",
	ObjectNameGroup:         "/groups",
	objectNameTrackingField: "/tracking_fields",
}, func(objectName string) (path string) {
	return objectName
},
)
