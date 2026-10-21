package orb

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	objectNameCustomers     = "customers"
	objectNameInvoices      = "invoices"
	objectNameMetrics       = "metrics"
	objectNameSubscriptions = "subscriptions"
)

var supportedObjects = []string{ // nolint:gochecknoglobals
	objectNameCustomers,
	objectNameInvoices,
	objectNameMetrics,
	objectNameSubscriptions,
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
