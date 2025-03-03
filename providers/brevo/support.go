package brevo

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	"blockedContacts",
	"categories",
	"children",
	"companies",
	"contacts",
	"deals",
	"emailCampaigns",
	"files",
	"folders",
	"inbound/events",
	"lists",
	"notes",
	"processes",
	"products",
	"smsCampaigns",
	"smtp/statistics/events",
	"smtp/statistics/reports",
	"subAccount",
	"tasks",
	"templates",
	"transactionalSMS/statistics/events",
)

func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
