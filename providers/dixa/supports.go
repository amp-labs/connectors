package dixa

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"agents", "knowledge/collections", "custom-attributes", "endusers",
		"conversations/flows", "analytics/metrics", "agents/presence", "queues",
		"analytics/records", "tags", "teams", "webhooks", "contact-endpoints",
		"business-hours/schedules", "templates",
	}

	writeSupport := []string{
		"agents", "agents/bulk", "conversations", "conversations/import", "endusers",
		"endusers/bulk", "queues", "tags", "teams", "webhooks",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			}, {
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
