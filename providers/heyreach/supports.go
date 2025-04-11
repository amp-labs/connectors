package heyreach

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/heyreach/metadata"
)

func supportedOperations() components.EndpointRegistryInput {
	supportWrite := []string{
		"list/CreateEmptyList",
		"campaign/AddLeadsToCampaignV2",
		"list/AddLeadsToListV2",
		"inbox/SendMessage",
	}

	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
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
