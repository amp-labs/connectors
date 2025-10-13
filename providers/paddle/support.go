package paddle

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/paddle/metadata"
)

//nolint:gochecknoglobals
var supportIncrementalRead = datautils.NewStringSet(
	"transactions",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	writeSupport := []string{
		"products",
		"prices",
		"discounts",
		"discount-groups",
		"customers",
		"transactions",
		"adjustments",
		"client-tokens",
		"reports",
		"notification-settings",
		"simulations",
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
		},
	}
}
