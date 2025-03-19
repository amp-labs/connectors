package brevo

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	"blockedContacts",
	"categories",
	"children",
	"companies",
	"contacts",
	"deals",
	"emailCampaigns",
	"files",
	"folders",
	"inbound/events",
	"lists",
	"notes",
	"processes",
	"products",
	"smsCampaigns",
	"smtp/statistics/events",
	"smtp/statistics/reports",
	"subAccount",
	"tasks",
	"templates",
	"transactionalSMS/statistics/events",
)

func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	writeSupport := []string{
		"smtp/email", "smtp/templates", "smtp/blockedDomains",
		"smtp/deleteHardbounces", "transactionalSMS/sms", "whatsapp/sendMessage",
		"emailCampaigns", "emailCampaigns/images", "smsCampaigns", "whatsappCampaigns",
		"whatsappCampaigns/template", "contacts", "contacts/doubleOptinConfirmation", "contacts/folders",
		"contacts/export", "contacts/import", "events", "senders", "senders/domains", "webhooks", "webhooks/export",
		"corporate/subAccount", "corporate/ssoToken", "corporate/subAccount/ssoToken", "corporate/subAccount/key",

		"corporate/group", "corporate/user/invitation/send",
		"organization/user/invitation/send", "organization/user/update/permissions",

		"feeds", "companies", "crm/attributes", "companies/import",
		"crm/deals", "crm/deals/import", "crm/tasks", "crm/notes",
		"conversations/messages",

		"conversations/pushedMessages", "conversations/agentOnlinePing",
		"ecommerce/activate", "ecommerce/config/displayCurrency",

		"orders/status", "orders/status/batch", "categories", "categories/batch", "products", "couponCollections",
		"coupons", "payments/requests", "loyalty/config/programs",
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
