package blueshift

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/blueshift/metadata"
)

const (
	objectNameCampaigns       = "campaigns"
	objectNameEmailTemplates  = "email_templates"
	objectNameExternalFetches = "external_fetches"
	objectNameOnsiteSlots     = "onsite_slots"
	objectNameSmsTemplates    = "sms_templates"
	objectNamePushTemplates   = "push_templates"
)

var nestedObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameEmailTemplates,
)

var supportPagination = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCampaigns,
	objectNameEmailTemplates,
	objectNameExternalFetches,
	objectNameOnsiteSlots,
	objectNameSmsTemplates,
	objectNamePushTemplates,
)

func supportedOperations() components.EndpointRegistryInput {
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
