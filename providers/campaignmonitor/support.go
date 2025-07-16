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
	writeSupport := []string{
		"admins",
		"clients",
	}

	deleteSupport := []string{
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
