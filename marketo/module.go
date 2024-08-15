package marketo

import "github.com/amp-labs/connectors/common/paramsbuilder"

// Assets is the module/API used for accessing assets objects.
var Assets = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "asset",
	Version: "v1",
}

// Leads is the module/API used for accessing leads objects.
var Leads = paramsbuilder.APIModule{ //nolint: gochecknoglobals
	Label:   "",
	Version: "v1",
}

// ModuleEmpty is used for proxying requests through.
var ModuleEmpty = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "",
	Version: "",
}

// supportedModules represents currently working and supported modules within the Hubspot connector.
// Any added module should be appended here.
var supportedModules = []paramsbuilder.APIModule{ // nolint: gochecknoglobals
	ModuleEmpty,
	Leads,
	Assets,
}
