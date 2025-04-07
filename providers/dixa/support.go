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

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
