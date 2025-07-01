package atlassian

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Deprecated.
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
}
