package getresponse

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
)

const (
	objectNameCampaigns = "campaigns"
	objectNameContacts  = "contacts"
)

var supportWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCampaigns,
	objectNameContacts,
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},

			// TODO: add write support when implemented
			// {
			// 	Endpoint: fmt.Sprintf("{%s}", strings.Join(supportWrite.List(), ",")),
			// 	Support:  components.WriteSupport,
			// },
		},
	}
}
