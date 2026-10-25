package orb

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const objectNameInvoices = "invoices"

var supportedObjects = []string{ // nolint:gochecknoglobals
	"customers",
	"subscriptions",
	objectNameInvoices,
	// Billable metrics, GET /v1/metrics. The list includes archived metrics (status "archived").
	"metrics",
}

// SupportedReadObjects lists the objects this connector reads.
func SupportedReadObjects() []string {
	return slices.Clone(supportedObjects)
}

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportedObjects, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
