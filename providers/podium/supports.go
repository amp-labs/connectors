package podium

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func supportsPagination(objectName string) bool {
	switch objectName {
	case "locations", "users", "campaign_interactions",
		"contacts", "conversations", "feedback", "invoices",
		"products", "reviews", "reviews/invites":
		return true
	default:
		return false
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"locations", "users", "campaign_interactions", "campaigns",
		"contact_attributes", "contact_tags", "contacts", "conversations",
		"feedback", "templates", "invoices", "products", "reviews", "reviews/invites", "reviews/sites/summary",
		"reviews/summary", "webhooks",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
