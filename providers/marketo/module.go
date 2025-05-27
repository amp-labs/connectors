package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleAssets
	// Deprecated.
	ModuleAssets = providers.ModuleMarketoAssets
)

// supportedModules represents currently working and supported modules within the Marketo connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "v1",
	},
	providers.ModuleMarketoAssets: {
		ID:      providers.ModuleMarketoAssets,
		Label:   "asset",
		Version: "v1",
	},
}
