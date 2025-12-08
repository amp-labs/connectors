package dropboxsign

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:gochecknoglobals
var (
	objectNameTemplate       = "template"
	objectNameBulkSendJobs   = "bulk_send_job"
	objectNameApiApp         = "api_app"
	objectNameFax            = "fax"
	objectNameFaxLine        = "fax_line"
	objectNameAccount        = "account"
	objectNameReport         = "report"
	objectNameTeam           = "team"
	objectNameUnclaimedDraft = "unclaimed_draft"
)

//nolint:gochecknoglobals
var readObjectResponseKey = datautils.NewDefaultMap(map[string]string{
	objectNameTemplate:     "templates",
	objectNameApiApp:       "api_apps",
	objectNameFax:          "faxes",
	objectNameFaxLine:      "fax_lines",
	objectNameBulkSendJobs: "bulk_send_jobs",
}, func(objectName string) (fieldName string) {
	return objectName
},
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		objectNameTemplate,
		objectNameApiApp,
		objectNameFax,
		objectNameFaxLine,
		objectNameBulkSendJobs,
	}

	writeSupport := []string{
		objectNameAccount,
		objectNameTemplate,
		objectNameReport,
		objectNameTeam,
		objectNameUnclaimedDraft,
		objectNameApiApp,
		objectNameFaxLine,
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
