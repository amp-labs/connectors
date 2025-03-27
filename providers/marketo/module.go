package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleAssets
	// Deprecated.
	ModuleAssets = common.ModuleID(providers.ModuleMarketoAssets)
	// ModuleLeads
	// Deprecated.
	ModuleLeads = common.ModuleID(providers.ModuleMarketoLeads)
)

// supportedModules represents currently working and supported modules within the Marketo connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	common.ModuleID(providers.ModuleMarketoAssets): {
		ID:      common.ModuleID(providers.ModuleMarketoAssets),
		Label:   "asset",
		Version: "v1",
	},
	common.ModuleID(providers.ModuleMarketoLeads): {
		ID:      common.ModuleID(providers.ModuleMarketoLeads),
		Label:   "",
		Version: "v1",
	},
}
