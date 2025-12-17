package shopify

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// records returns a function that extracts records from a GraphQL response.
// Shopify uses the connection pattern: data.{objectName}.nodes[]
func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node, "data", objectName).ArrayOptional("nodes")
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

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

		hasNextPage, err := jsonquery.New(pageInfo).BoolWithDefault("hasNextPage", false)
		if err != nil {
			return "", err
		}

		if !hasNextPage {
			return "", nil
		}

		return jsonquery.New(pageInfo).StrWithDefault("endCursor", "")
	}
}
