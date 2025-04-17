package heyreach

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/heyreach/metadata"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	supportWrite := []string{
		"list/CreateEmptyList",
		"campaign/AddLeadsToCampaignV2",
		"list/AddLeadsToListV2",
		"inbox/SendMessage",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportWrite, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
