package podium

import (
	"fmt"
	"net/http"
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

func updateMethod(objectName string) string {
	switch objectName {
	case "campaigns", "conversations", "templates", "webhooks":
		return http.MethodPut
	default:
		return http.MethodPatch
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"locations", "users", "campaign_interactions", "campaigns",
		"contact_attributes", "contact_tags", "contacts", "conversations",
		"feedback", "templates", "invoices", "products", "reviews", "reviews/invites", "reviews/sites/summary",
		"reviews/summary", "webhooks",
	}

	writeSupport := []string{
		"locations", "appointments", "campaigns", "contact_attributes",
		"contact_tags", "contacts", "conversations", "import/messages", "messages", "messages/attachment",
		"templates", "invoices", "refunds", "reviews/invites", "webhooks",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
