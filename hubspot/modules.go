package hubspot

import (
	"fmt"
)

type APIModule struct {
	Label   string // e.g. "crm"
	Version string // e.g. "v3"
}

var ModuleCRM = APIModule{ // nolint: gochecknoglobals
	Label:   "crm",
	Version: "v3",
}

func (a APIModule) String() string {
	return fmt.Sprintf("%s/%s/%s", a.Label, a.Version)
}
