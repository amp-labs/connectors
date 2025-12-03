package dropboxsign

import "github.com/amp-labs/connectors/internal/datautils"

//nolint:gochecknoglobals
var (
	objectNameTemplate     = "template"
	objectNameBulkSendJobs = "bulk_send_job"
	objectNameApiApp       = "api_app"
	objectNameFax          = "fax"
	objectNameFaxLine      = "fax_line"
)

//nolint:gochecknoglobals
var readObjectResponseKey = datautils.NewDefaultMap(map[string]string{
	objectNameTemplate: "templates",
	objectNameApiApp:   "api_apps",
	objectNameFax:      "faxes",
	objectNameFaxLine:  "fax_lines",
}, func(objectName string) (fieldName string) {
	return objectName
},
)
