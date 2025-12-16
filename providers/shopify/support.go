package shopify

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// supportedOperations returns the supported operations for the Shopify connector.
func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"products",
		"orders",
		"customers",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
