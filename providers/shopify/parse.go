package shopify

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// nextRecordsURL returns a function that extracts the next page cursor from the response.
// Shopify uses pageInfo with hasNextPage and endCursor for cursor-based pagination.
func nextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pageInfo, err := jsonquery.New(node, "data", objectName).ObjectOptional("pageInfo")
		if err != nil {
			return "", err
		}

		if pageInfo == nil {
			return "", nil
		}

		return jsonquery.New(pageInfo).StrWithDefault("endCursor", "")
	}
}
