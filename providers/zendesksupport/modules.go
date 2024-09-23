package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
)

const (
	// ModuleEmpty is used for proxying requests through.
	ModuleEmpty common.ModuleID = ""
	// ModuleTicketing is used for proxying requests through.
	// https://developer.zendesk.com/api-reference/ticketing/introduction/
	ModuleTicketing common.ModuleID = "ticketing"
	// ModuleHelpCenter is Zendesk Help Center.
	// https://developer.zendesk.com/api-reference/help_center/help-center-api/introduction/
	ModuleHelpCenter common.ModuleID = "help-center"
)

// SupportedModules represents currently working and supported modules within the Zendesk connector.
// Any added module should be appended here.
var SupportedModules = common.Modules{ // nolint: gochecknoglobals
	ModuleEmpty: {
		ID:      ModuleEmpty,
		Label:   "",
		Version: "",
	},
	ModuleTicketing: {
		ID:      ModuleTicketing,
		Label:   "",
		Version: "",
	},
	ModuleHelpCenter: {
		ID:      ModuleHelpCenter,
		Label:   "",
		Version: "",
	},
}
