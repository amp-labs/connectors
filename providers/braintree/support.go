package braintree

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// Braintree GraphQL API Documentation:
// https://developer.paypal.com/braintree/graphql/reference/
//
// Note: payment_methods cannot be searched independently in Braintree's GraphQL API.
// They must be accessed via a Customer's paymentMethods connection.
// See: https://developer.paypal.com/braintree/graphql/guides/payment_methods#search
func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"customers",
		"transactions",
		"refunds",
		"disputes",
		"verifications",
		"merchant_accounts",
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
