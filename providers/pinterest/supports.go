package pinterest

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"pins", "boards", "media", "ad_accounts", "catalogs", "employers", "feeds", "product_groups", "integrations", "stats",
	}

	writeSupport := []string{
		"pins", "boards", "media", "catalogs", "feeds", "websites", "ad_accounts", "product_groups", "commerce", "logs", "reports",
	}

	deleteSupport := []string{
		"pins", "boards", "commerce", "feeds", "product_groups",
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
