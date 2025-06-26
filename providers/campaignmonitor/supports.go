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
		"lists",
		"segments",
		"suppressionlist",
		"templates",
		"people",
		"tags",
		"campaigns",
		"scheduled",
		"drafts",
		"journeys",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
