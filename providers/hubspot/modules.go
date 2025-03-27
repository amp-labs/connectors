package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// ModuleCRM
	// Deprecated.
	ModuleCRM = common.ModuleID(providers.ModuleHubspotCRM)
)

// supportedModules represents currently working and supported modules within the Hubspot connector.
// Any added module should be appended added here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	common.ModuleRoot: {
		ID:      common.ModuleRoot,
		Label:   "",
		Version: "",
	},
	common.ModuleID(providers.ModuleHubspotCRM): {
		ID:      common.ModuleID(providers.ModuleHubspotCRM),
		Label:   "crm",
		Version: "v3",
	},
}
