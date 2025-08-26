package blackbaud

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	// Refer the link https://developer.blackbaud.com/skyapi/products/crm for read endpoints.
	readSupport := []string{
		"crm-adnmg/businessprocessstatus",
		"crm-adnmg/batchtemplates",
		"crm-adnmg/currencies",
		"crm-adnmg/sites",
		"crm-adnmg/businessprocessinstances",
		"crm-adnmg/businessprocessparameterset",
		"crm-evtmg/registrationtypes",
		"crm-evtmg/registrants",
		"crm-evtmg/events",
		"crm-evtmg/locations",
		"crm-fndmg/designations/hierarchies",
		"crm-fndmg/fundraisingpurposes",
		"crm-fndmg/educationalhistory",
		"crm-fndmg/fundraisingpurposetypes",
		"crm-fndmg/fundraisingpurposerecipients",
		"crm-mktmg/correspondencecodes",
		"crm-mktmg/appeals",
		"crm-mktmg/segments/recordsources",
		"crm-mktmg/solicitcodes",
		"crm-prsmg/prospectmanagers",
		"crm-prsmg/prospectopportunities",
		"crm-prsmg/prospects",
		"crm-prsmg/stewardshipplansteps",
		"crm-revmg/payments",
		"crm-revmg/revenuetransactions",
		"crm-volmg/volunteerassignments",
		"crm-volmg/occurrences",
		"crm-volmgjobs",
		"crm-volmg/volunteers",
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
