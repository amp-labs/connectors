package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleAssets
	// Deprecated.
	ModuleAssets = providers.ModuleMarketoAssets
	// ModuleLeads
	// Deprecated.
	ModuleLeads = providers.ModuleMarketoLeads
)

// supportedModules represents currently working and supported modules within the Marketo connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	providers.ModuleMarketoAssets: {
		ID:      providers.ModuleMarketoAssets,
		Label:   "asset",
		Version: "v1",
	},
	providers.ModuleMarketoLeads: {
		ID:      providers.ModuleMarketoLeads,
		Label:   "",
		Version: "v1",
	},
}
