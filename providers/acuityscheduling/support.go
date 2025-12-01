package acuityscheduling

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:gochecknoglobals
var supportObjectIncrementalRead = datautils.NewSet(
	"appointments",
	"availability/classes",
	"blocks",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"appointments",
		"appointment-addons",
		"appointment-types",
		"availability/classes",
		"blocks",
		"calendars",
		"certificates",
		"clients",
		"forms",
		"labels",
		"orders",
		"products",
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
