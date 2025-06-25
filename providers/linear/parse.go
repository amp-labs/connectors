package linear

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node, "data", objectName).ArrayOptional("nodes")
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		meta, err := jsonquery.New(node, "data", objectName).ObjectRequired("pageInfo")
		if err != nil {
			return "", err
		}

		nextPageExists, err := jsonquery.New(meta).BoolRequired("hasNextPage")
		if err != nil {
			return "", err
		}

		if !nextPageExists {
			return "", err
		}

		nextCursor, err := jsonquery.New(meta).StringRequired("endCursor")
		if err != nil {
			return "", err
		}

		return nextCursor, nil
	}
}
