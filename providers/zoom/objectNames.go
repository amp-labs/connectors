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
)

var supportedObjectsByWrite = map[common.ModuleID]datautils.StringSet{ // nolint: gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		ObjectNameContactGroup,
		ObjectNameUser,
		ObjectNameGroup,
		objectNameTrackingField,
		objectNameDevice,
		objectNameH322Device,
	),
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	ObjectNameContactGroup:  "/contacts/groups",
	ObjectNameUser:          "/users",
	ObjectNameGroup:         "/groups",
	objectNameTrackingField: "/tracking_fields",
	objectNameDevice:        "/devices",
	objectNameH322Device:    "/h323/devices",
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
