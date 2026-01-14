package braintree

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	braintreeVersionHeader = "Braintree-Version"
	braintreeVersion       = "2019-01-01" // Use a recent stable version
	// Braintree GraphQL API returns up to 50 results per page by default.
	// See: https://developer.paypal.com/braintree/graphql/guides/connections/
	defaultPageSize = 50

	// Object name constants.
	objectPaymentMethods   = "paymentMethods"
	objectMerchantAccounts = "merchantAccounts"
)

// Braintree GraphQL API Documentation:
// https://developer.paypal.com/braintree/graphql/reference/
//
// Braintree uses Relay-style cursor pagination for GraphQL queries.
// Objects are returned in a connection structure with edges and pageInfo.
//
// Note: paymentMethods cannot be searched independently in Braintree's GraphQL API.
// Payment methods must be accessed via a Customer's paymentMethods connection.

// makeNextRecordsURL creates a pagination function for Relay-style cursor pagination.
// Braintree uses the standard Relay connection pattern with pageInfo.
func makeNextRecordsURL(objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Navigate to pageInfo in the connection
		// merchantAccounts uses: { data: { viewer: { merchant: { merchantAccounts: { pageInfo: {...} } } } } }
		// Other objects use: { data: { search: { [objName]: { pageInfo: {...} } } } }
		var (
			pagination *ajson.Node
			err        error
		)

		if objName == objectMerchantAccounts {
			pagination, err = jsonquery.New(node, "data", "viewer", "merchant", objName).ObjectOptional("pageInfo")
		} else {
			pagination, err = jsonquery.New(node, "data", "search", objName).ObjectOptional("pageInfo")
		}

		if err != nil {
			return "", err
		}

		if pagination == nil {
			return "", nil
		}

		return jsonquery.New(pagination).StrWithDefault("endCursor", "")
	}
}
