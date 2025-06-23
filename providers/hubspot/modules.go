package hubspot

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
	providers.ModuleHubspotCRM: {
		ID:      providers.ModuleHubspotCRM,
		Label:   "crm",
		Version: "v3",
	},
}
