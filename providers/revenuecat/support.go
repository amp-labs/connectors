package revenuecat

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/revenuecat/metadata"
)

// supportedOperations declares which objects support read, write (create/update), and delete.
// Docs: https://www.revenuecat.com/docs/api-v2
func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	//nolint:lll
	writeSupport := []string{
		// https://www.revenuecat.com/docs/api-v2#tag/Project-Configuration/operation/create-app
		"apps",
		// https://www.revenuecat.com/docs/api-v2#tag/Customer/operation/create-customer
		// No generic update endpoint; attributes are set via a sub-path.
		"customers",
		// https://www.revenuecat.com/docs/api-v2#tag/Entitlement/operation/create-entitlement
		"entitlements",
		// https://www.revenuecat.com/docs/api-v2#tag/Integration/operation/create-webhook-integration
		"integrations_webhooks",
		// https://www.revenuecat.com/docs/api-v2#tag/Offering/operation/create-offering
		"offerings",
		// https://www.revenuecat.com/docs/api-v2#tag/Product/operation/create-product
		// No update endpoint; create and delete only.
		"products",
	}

	//nolint:lll
	deleteSupport := []string{
		// https://www.revenuecat.com/docs/api-v2#tag/Project-Configuration/operation/delete-app
		"apps",
		// https://www.revenuecat.com/docs/api-v2#tag/Customer/operation/delete-customer
		"customers",
		// https://www.revenuecat.com/docs/api-v2#tag/Entitlement/operation/delete-entitlement
		"entitlements",
		// https://www.revenuecat.com/docs/api-v2#tag/Integration/operation/delete-webhook-integration
		"integrations_webhooks",
		// https://www.revenuecat.com/docs/api-v2#tag/Offering/operation/delete-offering
		"offerings",
		// https://www.revenuecat.com/docs/api-v2#tag/Product/operation/delete-product
		"products",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
