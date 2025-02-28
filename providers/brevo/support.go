package brevo

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func supportedOperations() components.EndpointRegistryInput {
	writeSupport := []string{ //nolint:lll
		"smtp/email", "smtp/templates", "smtp/blockedDomains",
		"smtp/deleteHardbounces", "transactionalSMS/sms", "whatsapp/sendMessage",
		"emailCampaigns", "emailCampaigns/images", "smsCampaigns", "whatsappCampaigns",
		"whatsappCampaigns/template", "contacts", "contacts/doubleOptinConfirmation", "contacts/folders",
		"contacts/export", "contacts/import", "events", "senders", "senders/domains", "webhooks", "webhooks/export",
		"corporate/subAccount", "corporate/ssoToken", "corporate/subAccount/ssoToken", "corporate/subAccount/key",
		"corporate/group", "corporate/user/invitation/send", "organization/user/invitation/send", "organization/user/update/permissions",
		"feeds", "companies", "crm/attributes", "companies/import", "crm/deals", "crm/deals/import", "crm/tasks", "crm/notes", "conversations/messages",
		"conversations/pushedMessages", "conversations/agentOnlinePing", "ecommerce/activate", "ecommerce/config/displayCurrency",
		"orders/status", "orders/status/batch", "categories", "categories/batch", "products", "couponCollections",
		"coupons", "payments/requests", "loyalty/config/programs",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
