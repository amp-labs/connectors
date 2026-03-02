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

	// customers: create + delete only (no generic update endpoint; attributes use a sub-path).
	// products:  create + delete only (no update endpoint).
	// apps, entitlements, offerings, integrations_webhooks: full create/update/delete.
	writeSupport := []string{
		"apps",
		"customers",
		"entitlements",
		"integrations_webhooks",
		"offerings",
		"products",
	}

	deleteSupport := writeSupport

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
