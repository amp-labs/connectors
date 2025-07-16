package campaignmonitor

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"clients",
		"admins",
	}

	// List of supported write endpoints based on official Campaign Monitor API documentation:
	// - https://www.campaignmonitor.com/api/v3-3/account/#adding-an-administrator.
	// - https://www.campaignmonitor.com/api/v3-3/clients/#creating-a-client.
	// - https://www.campaignmonitor.com/api/v3-3/campaigns/#creating-draft-campaign.
	// - https://www.campaignmonitor.com/api/v3-3/templates/#creating-a-template.
	// - https://www.campaignmonitor.com/api/v3-3/lists/#creating-a-list.
	// - https://www.campaignmonitor.com/api/v3-3/clients/#suppress-email-addresses.
	// - https://www.campaignmonitor.com/api/v3-3/clients/#transfer-credits.
	// - https://www.campaignmonitor.com/api/v3-3/clients/#adding-a-person.
	// - https://www.campaignmonitor.com/api/v3-3/clients/#add-a-sending-domain-client-specifies-the-keys-and-selector-2.
	writeSupport := []string{
		"admins",
		"clients",
		"campaigns",
		"templates",
		"lists",
		"suppress",
		"credits",
		"people",
		"sendingdomains",
	}

	deleteSupport := []string{
		"campaigns",
		"templates",
		"lists",
		"clients",
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
