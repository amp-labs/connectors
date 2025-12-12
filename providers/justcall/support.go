package justcall

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
)

// writableObjects lists objects that support write operations.
// Note: Some objects require additional setup (10DLC, WhatsApp Business, AI agents).
// Reference: https://developer.justcall.io/reference/introduction
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
// Note: webhooks use dedicated subscriber connector pattern for production use.
var deletableObjects = []string{ //nolint:gochecknoglobals
	"contacts",
	"tags",
	"sales_dialer/contacts",
	"webhooks",
}

// supportedOperations returns the endpoint registry for JustCall connector.
// Includes read, write, and delete support for various objects.
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
