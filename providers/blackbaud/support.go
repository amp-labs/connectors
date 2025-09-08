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
		// The following endpoints are available in CRM Administration.
		// Ref https://developer.sky.blackbaud.com/api#api=crm-adnmg.
		"crm-adnmg/batchtemplates",
		"crm-adnmg/businessprocessinstances",
		"crm-adnmg/businessprocessparameterset",
		"crm-adnmg/businessprocessstatus",
		"crm-adnmg/currencies",
		"crm-adnmg/sites",

		// The following endpoints are available in CRM Event.
		// https://developer.sky.blackbaud.com/api#api=crm-evtmg.
		"crm-evtmg/events",
		"crm-evtmg/locations",
		"crm-evtmg/registrants",
		"crm-evtmg/registrationtypes",

		// The following endpoints are available in CRM Fundraising.
		// https://developer.sky.blackbaud.com/api#api=crm-fndmg.
		"crm-fndmg/designations/hierarchies",
		"crm-fndmg/educationalhistory",
		"crm-fndmg/fundraisingpurposerecipients",
		"crm-fndmg/fundraisingpurposes",
		"crm-fndmg/fundraisingpurposetypes",

		// The following endpoints are available in CRM Marketing.
		// https://developer.sky.blackbaud.com/api#api=crm-mktmg.
		"crm-mktmg/appeals",
		"crm-mktmg/correspondencecodes",
		"crm-mktmg/segments/recordsources",
		"crm-mktmg/solicitcodes",

		// The following endpoints are available in CRM Prospect.
		// https://developer.sky.blackbaud.com/api#api=crm-prsmg.
		"crm-prsmg/prospectmanagers",
		"crm-prsmg/prospectopportunities",
		"crm-prsmg/prospects",
		"crm-prsmg/stewardshipplansteps",

		// The following endpoints are available in CRM Revenue.
		// https://developer.sky.blackbaud.com/api#api=crm-revmg.
		"crm-revmg/payments",
		"crm-revmg/revenuetransactions",

		// The following endpoints are available in CRM Volunteer.
		// https://developer.sky.blackbaud.com/api#api=crm-volmg.
		"crm-volmg/jobs",
		"crm-volmg/occurrences",
		"crm-volmg/volunteerassignments",
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
