package justcall

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
)

// writableObjects lists objects that support write operations.
var writableObjects = []string{ //nolint:gochecknoglobals
	"contacts",
	"contacts/status",
	"tags",
	"users/availability",
	"calls",
	"texts",
	"texts/threads/tag",
	"sales_dialer/campaigns/contact",
	"voice-agents/calls",
}

// deletableObjects lists objects that support delete operations.
var deletableObjects = []string{ //nolint:gochecknoglobals
	"contacts",
	"tags",
	"sales_dialer/contacts",
	"webhooks",
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writableObjects, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deletableObjects, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
