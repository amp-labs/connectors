package hubspot

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
)

// ModuleCRM is the module used for accessing standard CRM objects.
var ModuleCRM = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "crm",
	Version: "v3",
}

// ModuleEmpty is used for proxying requests through.
var ModuleEmpty = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "",
	Version: "",
}

// supportedModules represents currently working and supported modules within the Hubspot connector.
// Any added module should be appended added here.
var supportedModules = []paramsbuilder.APIModule{ // nolint: gochecknoglobals
	ModuleEmpty,
	ModuleCRM,
}
