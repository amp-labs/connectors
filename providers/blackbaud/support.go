package blackbaud

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// nolint:funlen
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

	writeSupport := []string{
		// The following endpoints are available in CRM Administration.
		// Ref https://developer.sky.blackbaud.com/api#api=crm-adnmg.
		"crm-adnmg/batches",
		"crm-adnmg/batches/revenue",
		"crm-adnmg/businessprocess/launch",
		"crm-adnmg/notifications",

		// The following endpoints are available in CRM Constituent.
		// Ref https://developer.sky.blackbaud.com/api#api=crm-conmg.
		"crm-conmg/addresses",
		"crm-conmg/alternatelookupids",
		"crm-conmg/constituentappealresponses",
		"crm-conmg/constituentappeals",
		"crm-conmg/constituentattributes",
		"crm-conmg/constituentcorrespondencecodes",
		"crm-conmg/constituentnotes",
		"crm-conmg/constituents",
		"crm-conmg/educationalhistories",
		"crm-conmg/emailaddresses",
		"crm-conmg/fundraisers",
		"crm-conmg/individuals",
		"crm-conmg/interaction",
		"crm-conmg/mergetwoconstituents",
		"crm-conmg/organizations",
		"crm-conmg/phones",
		"crm-conmg/relationshipjobsinfo",
		"crm-conmg/solicitcodes",
		"crm-conmg/tribute",

		// The following endpoints are available in CRM Event.
		// https://developer.sky.blackbaud.com/api#api=crm-evtmg.
		"crm-evtmg/events",
		"crm-evtmg/locations",
		"crm-evtmg/registrants",
		"crm-evtmg/registrationoptions",
		"crm-evtmg/registrationtypes",

		// The following endpoints are available in CRM Fundraising.
		// https://developer.sky.blackbaud.com/api#api=crm-fndmg.
		"crm-fndmg/fundraisingpurposerecipients",
		"crm-fndmg/fundraisingpurposes",

		// The following endpoints are available in CRM Marketing.
		// https://developer.sky.blackbaud.com/api#api=crm-mktmg.
		"crm-mktmg/appeals",
		"crm-mktmg/correspondencecodes",
		"crm-mktmg/responsecategories",
		"crm-mktmg/segments",

		// The following endpoints are available in CRM Prospect.
		// https://developer.sky.blackbaud.com/api#api=crm-prsmg.
		"crm-prsmg/prospectcontactreports",
		"crm-prsmg/prospectopportunities",
		"crm-prsmg/prospectplans",
		"crm-prsmg/prospectsegmentations",
		"crm-prsmg/prospects",
		"crm-prsmg/prospectsconstituency",
		"crm-prsmg/prospectsteps",
		"crm-prsmg/stewardshipplans",
		"crm-prsmg/stewardshipplansteps",

		// The following endpoints are available in CRM Revenue.
		// https://developer.sky.blackbaud.com/api#api=crm-revmg.
		"crm-revmg/payments",
		"crm-revmg/recurringgifts",
		"crm-revmg/revenuenotes",

		// The following endpoints are available in CRM Volunteer.
		// https://developer.sky.blackbaud.com/api#api=crm-volmg.
		"crm-volmg/jobs",
		"crm-volmg/occurrences",
		"crm-volmg/timesheets",
		"crm-volmg/volunteerassignments",
		"crm-volmg/volunteers",
		"crm-volmg/volunteerschedules",
	}

	deleteSupport := []string{
		// The following endpoints are available in CRM Constituent.
		// Ref https://developer.sky.blackbaud.com/api#api=crm-conmg.
		"crm-conmg/addresses",
		"crm-conmg/alternatelookupids",
		"crm-conmg/constituentappeals",
		"crm-conmg/constituentattributes",
		"crm-conmg/constituentcorrespondencecodes",
		"crm-conmg/constituentnotes",
		"crm-conmg/educationalhistories",
		"crm-conmg/emailaddresses",
		"crm-conmg/fundraisers",
		"crm-conmg/interaction",
		"crm-conmg/phones",
		"crm-conmg/relationshipjobsinfo",
		"crm-conmg/solicitcodes",
		"crm-conmg/tribute",

		// The following endpoints are available in CRM Event.
		// https://developer.sky.blackbaud.com/api#api=crm-evtmg.
		"crm-evtmg/events",
		"crm-evtmg/locations",
		"crm-evtmg/registrants",
		"crm-evtmg/registrationoptions",
		"crm-evtmg/registrationtypes",

		// The following endpoints are available in CRM Fundraising.
		// https://developer.sky.blackbaud.com/api#api=crm-fndmg.
		"crm-fndmg/fundraisingpurposes",
		"crm-fndmg/fundraisingpurposerecipients",

		// The following endpoints are available in CRM Marketing.
		// https://developer.sky.blackbaud.com/api#api=crm-mktmg.
		"crm-mktmg/appeals",
		"crm-mktmg/correspondencecodes",
		"crm-mktmg/responsecategories",
		"crm-mktmg/segments",

		// The following endpoints are available in CRM Prospect.
		// https://developer.sky.blackbaud.com/api#api=crm-prsmg.
		"crm-prsmg/prospectopportunities",
		"crm-prsmg/prospectplans",
		"crm-prsmg/prospectsegmentations",
		"crm-prsmg/prospectsteps",
		"crm-prsmg/stewardshipplans",
		"crm-prsmg/stewardshipplansteps",

		// The following endpoints are available in CRM Revenue.
		// https://developer.sky.blackbaud.com/api#api=crm-revmg.
		"crm-revmg/payments",

		// The following endpoints are available in CRM Volunteer.
		// https://developer.sky.blackbaud.com/api#api=crm-volmg.
		"crm-volmg/jobs",
		"crm-volmg/occurrences",
		"crm-volmg/timesheets",
		"crm-volmg/volunteerassignments",
		"crm-volmg/volunteers",
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
