package gitlab

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		switch objectResponders.Has(objectName) {
		case true:
			record, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return nil, err
			}

			return []map[string]any{record}, nil
		default:
			data, err := jsonquery.New(node).ArrayOptional("")
			if err != nil {
				return nil, err
			}

			return jsonquery.Convertor.ArrayToMap(data)
		}
	}
}

func nextRecordsURL(nextPage string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return nextPage, nil
	}
}
