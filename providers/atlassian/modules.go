package atlassian

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleJira
	// Deprecated.
	ModuleJira = common.ModuleID(providers.ModuleAtlassianJira)
	// ModuleAtlassianJiraConnect
	// Deprecated.
	ModuleAtlassianJiraConnect = common.ModuleID(providers.ModuleAtlassianJiraConnect)
)

// supportedModules represents currently working and supported modules within the Atlassian connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	common.ModuleID(providers.ModuleAtlassianJira): {
		ID:      common.ModuleID(providers.ModuleAtlassianJira),
		Label:   "rest/api",
		Version: "3",
	},
	common.ModuleID(providers.ModuleAtlassianJiraConnect): {
		ID:      common.ModuleID(providers.ModuleAtlassianJiraConnect),
		Label:   "rest/api",
		Version: "3",
	},
}
