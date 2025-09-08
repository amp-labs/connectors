package blackbaud

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 1

var objectNameWithSearchResource = datautils.NewSet( //nolint:gochecknoglobals
	"crm-adnmg/sites",
	"crm-adnmg/businessprocessparameterset",
	"crm-evtmg/registrationtypes",
	"crm-evtmg/registrants",
	"crm-evtmg/events",
	"crm-fndmg/fundraisingpurposes",
	"crm-fndmg/educationalhistory",
	"crm-fndmg/fundraisingpurposerecipients",
	"crm-prsmg/prospectmanagers",
	"crm-prsmg/prospectopportunities",
	"crm-prsmg/prospects",
	"crm-prsmg/stewardshipplansteps",
	"crm-revmg/payments",
	"crm-revmg/revenuetransactions",
	"crm-volmg/volunteerassignments",
	"crm-volmg/occurrences",
	"crm-volmg/jobs",
	"crm-volmg/volunteers",
)

var objectNameWithListResource = datautils.NewSet( //nolint:gochecknoglobals
	"crm-adnmg/businessprocessstatus",
	"crm-adnmg/batchtemplates",
	"crm-adnmg/currencies",
	"crm-adnmg/businessprocessinstances",
	"crm-evtmg/locations",
	"crm-fndmg/designations/hierarchies",
	"crm-mktmg/correspondencecodes",
	"crm-mktmg/appeals",
	"crm-mktmg/solicitcodes",
)

func makeNextRecord() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return "", nil
	}
}
