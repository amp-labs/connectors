package atlassian

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
)

// ModuleJira is the module used for listing Jira issues.
var ModuleJira = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "rest/api",
	Version: "3",
}

// ModuleEmpty is used for proxying requests through.
var ModuleEmpty = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "",
	Version: "",
}

// supportedModules represents currently working and supported modules within the Atlassian connector.
// Any added module should be appended here.
var supportedModules = []paramsbuilder.APIModule{ // nolint: gochecknoglobals
	ModuleEmpty,
	ModuleJira,
}
