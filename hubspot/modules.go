package hubspot

import (
	"fmt"
)

type APIModule struct {
	Label      string // e.g. "crm"
	Version    string // e.g. "v3"
	SuffixBase string // e.g. "objects"
}

var ModuleCRM = APIModule{ // nolint: gochecknoglobals
	Label:      "crm",
	Version:    "v3",
	SuffixBase: "objects",
}

func (a APIModule) String() string {
	return fmt.Sprintf("%s/%s/%s", a.Label, a.Version, a.SuffixBase)
}
