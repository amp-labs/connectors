package servicenow

import "github.com/amp-labs/connectors/common"

const (
	// ModuleTable is the module/API used for accessing and manging table records.
	ModuleTable common.ModuleID = "table"
)

// supportedModules represents currently working and supported modules within the Marketo connector.
// Any added module should be appended here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	ModuleTable: {
		ID:      ModuleTable,
		Label:   "now",
		Version: "v2/table",
	},
	// Add other modules.
}
