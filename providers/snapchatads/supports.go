package snapchatads

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"fundingsources",
		"billingcenters",
		"transactions",
		"adaccounts",
		"members",
		"roles",
		"age_group",
		"gender",
		"languages",
		"advanced_demographics",
		"connection_type",
		"os_type",
		"carrier",
		"marketing_name",
		"country",
		"dlxs",
		"dlxp",
		"nln",
		"categories_loi",
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
