package hubspot

import (
	"fmt"
)

type APIModule struct {
	Label   string // e.g. "crm"
	Version string // e.g. "v3"
}

// ModuleCRM is the module used for accessing standard CRM objects.
var ModuleCRM = APIModule{ // nolint: gochecknoglobals
	Label:   "crm",
	Version: "v3",
}

// ModuleEmpty is used for proxying requests through.
var ModuleEmpty = APIModule{ // nolint: gochecknoglobals
	Label:   "",
	Version: "",
}

// supportedModules represents currently working and supported modules within the Hubspot connector.
// Any added module should be appended added here.
var supportedModules = []APIModule{ModuleCRM} // nolint: gochecknoglobals

func (a APIModule) String() string {
	return fmt.Sprintf("%s/%s", a.Label, a.Version)
}

func supportsModule(module string) bool {
	for _, mod := range supportedModules {
		if module == mod.String() {
			return true
		}
	}

	return false
}
