package braintree

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	braintreeVersionHeader = "Braintree-Version"
	braintreeVersion       = "2019-01-01" // Use a recent stable version
	// Braintree GraphQL API returns up to 50 results per page by default.
	// See: https://developer.paypal.com/braintree/graphql/guides/connections/
	defaultPageSize = 50
)

// Braintree GraphQL API Documentation:
// https://developer.paypal.com/braintree/graphql/reference/
//
// Braintree uses Relay-style cursor pagination for GraphQL queries.
// Objects are returned in a connection structure with edges and pageInfo.

// objectNameToGraphQLField maps connector object names (snake_case) to GraphQL response field names (camelCase).
// This is needed because GraphQL responses use camelCase while connector uses snake_case.
// Note: payment_methods is not included here as it cannot be searched independently.
// Payment methods must be accessed via a Customer's paymentMethods connection.
var objectNameToGraphQLField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"customers":         "customers",
	"transactions":      "transactions",
	"refunds":           "refunds",
	"disputes":          "disputes",
	"verifications":     "verifications",
	"merchant_accounts": "merchantAccounts",
}, snakeToCamel)

// snakeToCamel converts snake_case to camelCase.
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	result := parts[0]

	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// makeNextRecordsURL creates a pagination function for Relay-style cursor pagination.
// Braintree uses the standard Relay connection pattern with pageInfo.
func makeNextRecordsURL(objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		graphqlFieldName := objectNameToGraphQLField.Get(objName)

		// Navigate to pageInfo in the connection
		// merchant_accounts uses: { data: { viewer: { merchant: { merchantAccounts: { pageInfo: {...} } } } } }
		// Other objects use: { data: { search: { [objName]: { pageInfo: {...} } } } }
		var (
			pagination *ajson.Node
			err        error
		)

		if objName == "merchant_accounts" {
			pagination, err = jsonquery.New(node, "data", "viewer", "merchant", graphqlFieldName).ObjectOptional("pageInfo")
		} else {
			pagination, err = jsonquery.New(node, "data", "search", graphqlFieldName).ObjectOptional("pageInfo")
		}

		if err != nil {
			return "", err
		}

		if pagination == nil {
			return "", nil
		}

		hasNextPage, err := jsonquery.New(pagination).BoolOptional("hasNextPage")
		if err != nil {
			return "", err
		}

		if hasNextPage == nil || !(*hasNextPage) {
			return "", nil
		}

		endCursor, err := jsonquery.New(pagination).StringOptional("endCursor")
		if err != nil {
			return "", err
		}

		if endCursor == nil {
			return "", nil
		}

		return *endCursor, nil
	}
}
