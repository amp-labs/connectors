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

// ModuleEmpty is Used for proxying requests through.
var ModuleEmpty = APIModule{ // nolint: gochecknoglobals
	Label:   "",
	Version: "",
}

func (a APIModule) String() string {
	return fmt.Sprintf("%s/%s", a.Label, a.Version)
}
