package atlassian

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleJira
	// Deprecated.
	ModuleJira = providers.ModuleAtlassianJira
	// ModuleAtlassianJiraConnect
	// Deprecated.
	ModuleAtlassianJiraConnect = providers.ModuleAtlassianJiraConnect
)

// supportedModules represents currently working and supported modules within the Atlassian connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	providers.ModuleAtlassianJira: {
		ID:      providers.ModuleAtlassianJira,
		Label:   "rest/api",
		Version: "3",
	},
	providers.ModuleAtlassianJiraConnect: {
		ID:      providers.ModuleAtlassianJiraConnect,
		Label:   "rest/api",
		Version: "3",
	},
}
