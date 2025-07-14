package calendly

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: "scheduled_events",
				Support:  components.ReadSupport,
			},
			{
				Endpoint: "webhook_subscriptions",
				Support:  providers.Support{Subscribe: true},
			},
		},
	}
} 