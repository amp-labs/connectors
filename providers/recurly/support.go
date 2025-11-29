package recurly

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/recurly/metadata"
)

//nolint:gochecknoglobals
var supportIncrementalRead = datautils.NewStringSet(
	"accounts",
	"acquisitions",
	"subscriptions",
	"items",
	"plans",
	"add_ons",
	"measured_units",
	"coupons",
	"invoices",
	"line_items",
	"credit_payments",
	"transactions",
	"custom_field_definitions",
	"shipping_methods",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	writeSupport := []string{
		"accounts",
		"coupons",
		"external_products",
		"external_subscriptions",
		"general_ledger_accounts",
		"gift_cards",
		"items",
		"measured_units",
		"plans",
		"shipping_methods",
		"subscriptions",
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
