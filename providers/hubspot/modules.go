package hubspot

import (
	"github.com/amp-labs/connectors/common"
)

const (
	// ModuleEmpty is used for proxying requests through.
	ModuleEmpty common.ModuleID = ""
	// ModuleCRM is the module used for accessing standard CRM objects.
	ModuleCRM common.ModuleID = "CRM"
)

// supportedModules represents currently working and supported modules within the Hubspot connector.
// Any added module should be appended added here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	ModuleEmpty: {
		ID:      ModuleEmpty,
		Label:   "",
		Version: "",
	},
	ModuleCRM: {
		ID:      ModuleCRM,
		Label:   "crm",
		Version: "v3",
	},
}
