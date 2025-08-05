package highlevelstandard

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// nolint:funlen
func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"businesses",
		"calendars",
		"calendars/groups",
		"campaigns",
		"conversations/search",
		"emails/schedule",
		"forms/submissions",
		"forms",
		"invoices",
		"invoices/template",
		"invoices/schedule",
		"invoices/estimate/list",
		"invoices/estimate/template",
		"links",
		"blogs/authors",
		"blogs/categories",
		"funnels/lookup/redirect/list",
		"funnels/funnel/list",
		"opportunities/pipelines",
		"payment/orders",
		"payments/transactions",
		"payments/subscriptions",
		"payments/coupon/list",
		"products",
		"products/inventory",
		"products/collections",
		"products/reviews",
		"proposals/document",
		"proposals/templates",
		"store/shipping-zone",
		"store/shipping-carrier",
		"store/store-setting",
		"snapshots",
		"surveys",
		"users",
		"workflows",
		"locations/search",
		"custom-menus",
	}

	writeSupport := []string{
		"custom-menus",
		"users",
		"businesses",
		"calendars",
		"calendars/groups",
		"calendars/events/appointments",
		"calendars/events/block-slots",
		"contacts",
		"objects",
		"associations",
		"associations/relations",
		"custom-fields",
		"custom-fields/folder",
		"conversations",
		"conversations/messages",
		"conversations/messages/inbound",
		"conversations/messages/outbound",
		"conversations/messages/upload",
		"emails/builder",
		"invoices",
		"invoices/template",
		"invoices/schedule",
		"invoices/text2pay",
		"invoices/estimate",
		"invoices/estimate/template",
		"links",
		"locations",
		"blogs/posts",
		"funnels/lookup/redirect",
		"opportunities",
		"payments/coupon",
		"products",
		"products/collections",
		"store/shipping-zone",
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
		},
	}
}
