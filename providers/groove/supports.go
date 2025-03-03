package groove

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

var responseField = map[string]string{ //nolint:gochecknoglobals
	"tickets":       "tickets",
	"customers":     "customers",
	"mailboxes":     "mailboxes",
	"folders":       "folders",
	"agents":        "agents",
	"groups":        "groups",
	"kb/themes":     "themes",
	"widgets":       "widgets",
	"kb":            "knowledge_bases",
	"tickets/count": "",
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"tickets", "customers", "mailboxes", "folders",
		"agents", "groups", "kb/themes", "widgets",
		"kb", "tickets/count",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
