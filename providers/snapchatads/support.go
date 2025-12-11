package snapchatads

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	// readSupport lists all endpoints that can be fetched for the default organization.
	readSupport := []string{
		"fundingsources",
		"billingcenters",
		"transactions",
		"adaccounts",
		"members",
		"roles",
		"targeting/demographics/age_group",
		"targeting/demographics/gender",
		"targeting/demographics/languages",
		"targeting/demographics/advanced_demographics",
		"targeting/device/connection_type",
		"targeting/device/os_type",
		"targeting/device/carrier",
		"targeting/device/marketing_name",
		"targeting/geo/country",
		"targeting/interests/dlxs",
		"targeting/interests/dlxc",
		"targeting/interests/dlxp",
		"targeting/interests/nln",
		"targeting/location/categories_loi",
	}

	writeSupport := []string{
		"billingcenters",
		"adaccounts",
		"members",
		"roles",
	}

	deleteSupport := []string{
		"members",
		"roles",
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
