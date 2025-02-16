package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

var supportedObjectsByRead = metadata.Schemas.ObjectNames() // nolint: gochecknoglobals

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
	ModuleUser: datautils.NewDefaultMap(map[string]string{},
		func(objectName string) (fieldName string) {
			return objectName
		},
	),
}
