package dixa

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"agents", "knowledge/collections", "custom-attributes", "endusers",
		"conversations/flows", "analytics/metrics", "agents/presence", "queues",
		"analytics/records", "tags", "teams", "webhooks", "contact-endpoints",
		"business-hours/schedules", "templates",
	}

	writeSupport := []string{
		"agents", "conversations", "conversations/import", "endusers",
		"queues", "tags", "teams", "webhooks",
		// "agents/bulk","endusers/bulk" supports bulk write.
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
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
