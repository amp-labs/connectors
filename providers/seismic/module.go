package seismic

import "github.com/amp-labs/connectors/common"

const (
	// ModuleReporting is the module used for accessing and manging reporting API.
	ModuleReporting common.ModuleID = "reporting"
)

var modules = common.RequireModule{ //nolint:gochecknoglobals
	ExpectedModules: []common.ModuleID{ModuleReporting},
}
