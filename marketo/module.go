package marketo

import "github.com/amp-labs/connectors/common/paramsbuilder"

// ModuleAssets is the module/API used for accessing assets objects.
var ModuleAssets = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "asset",
	Version: "v1",
}

// ModuleLeads is the module/API used for accessing leads objects.
var ModuleLeads = paramsbuilder.APIModule{ //nolint: gochecknoglobals
	Label:   "",
	Version: "v1",
}

// supportedModules represents currently working and supported modules within the Marketo connector.
// Any added module should be appended here.
var supportedModules = []paramsbuilder.APIModule{ // nolint: gochecknoglobals
	ModuleLeads,
	ModuleAssets,
}
