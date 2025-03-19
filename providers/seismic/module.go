package seismic

import "github.com/amp-labs/connectors/common"

const (
	// ModuleReporting is the module used for accessing and manging reporting API.
	ModuleReporting common.ModuleID = "reporting"
)

// supportedModules represents currently working and supported modules within the Seismic connector.
// Any added module should be added here.
var supportedModules = common.Modules{ // nolint: gochecknoglobals
	ModuleReporting: {
		ID:      ModuleReporting,
		Label:   "reporting",
		Version: "v2",
	},
}

var modules = common.RequireModule{ //nolint:gochecknoglobals
	ExpectedModules: []common.ModuleID{ModuleReporting},
}
