package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleCRM
	// Deprecated.
	ModuleCRM = providers.ModuleHubspotCRM
)

// supportedModules represents currently working and supported modules within the Hubspot connector.
// Any added module should be appended added here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	providers.ModuleHubspotCRM: {
		ID:      providers.ModuleHubspotCRM,
		Label:   "crm",
		Version: "v3",
	},
}
