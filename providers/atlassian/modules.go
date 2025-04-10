package atlassian

import (
	"github.com/amp-labs/connectors/common"
)

const (
	// ModuleJira is the module used for listing Jira issues.
	ModuleJira common.ModuleID = "jira"
	// ModuleAtlassianJiraConnect is the module used for Atlassian Connect.
	ModuleAtlassianJiraConnect common.ModuleID = "atlassian-connect"
)

// supportedModules represents currently working and supported modules within the Atlassian connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	ModuleJira: {
		ID:      ModuleJira,
		Label:   "rest/api",
		Version: "3",
	},
	ModuleAtlassianJiraConnect: {
		ID:      ModuleAtlassianJiraConnect,
		Label:   "rest/api",
		Version: "3",
	},
}
