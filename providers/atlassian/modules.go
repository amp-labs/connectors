package atlassian

import (
	"github.com/amp-labs/connectors/common"
)

const (
	// ModuleEmpty is used for proxying requests through.
	ModuleEmpty common.ModuleID = ""
	// ModuleJira is the module used for listing Jira issues.
	ModuleJira common.ModuleID = "jira"
)

// supportedModules represents currently working and supported modules within the Atlassian connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	ModuleEmpty: {
		ID:      ModuleEmpty,
		Label:   "",
		Version: "",
	},
	ModuleJira: {
		ID:      ModuleJira,
		Label:   "rest/api",
		Version: "3",
	},
}
