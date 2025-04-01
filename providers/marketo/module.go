package marketo

import (
	"github.com/amp-labs/connectors/common"
)

const (
	// ModuleAssets is the module/API used for accessing assets objects.
	ModuleAssets common.ModuleID = "assets"
	// ModuleLeads is the module/API used for accessing leads objects.
	ModuleLeads common.ModuleID = "leads"
)

// supportedModules represents currently working and supported modules within the Marketo connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	ModuleAssets: {
		ID:      ModuleAssets,
		Label:   "asset",
		Version: "v1",
	},
	ModuleLeads: {
		ID:      ModuleLeads,
		Label:   "",
		Version: "v1",
	},
}
