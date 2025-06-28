package lever

import (
	"fmt"

	"github.com/amp-labs/connectors/internal/datautils"
)

const apiVersion = "v1"

var EndpointWithOpportunityID = datautils.NewSet( //nolint:gochecknoglobals
	"feedback",
	"files",
	"interviews",
	"notes",
	"offers",
	"panels",
	"forms",
	"referrals",
	"resumes",
)

func (c *Connector) constructURL(objName string) string {
	if EndpointWithOpportunityID.Has(objName) {
		return fmt.Sprintf("opportunities/%s/%s", c.opportunityId, objName)
	}

	return objName
}
