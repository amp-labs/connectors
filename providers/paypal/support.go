package paypal

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readObjects := []string{
		"balances",
		"disputes",
		"invoices",
		"plans",
		"products",
		"templates",
		"transactions",
		"web-profiles",
		"webhooks",
		"webhooks-event-types",
		"webhooks-events",
	}

	writeObjects := []string{
		"invoices",
		"orders",
		"plans",
		"products",
		"templates",
		"web-profiles",
		"webhooks",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readObjects, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeObjects, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
