package slack

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		responseKey := objectResponseField.Get(objectName)

		arr, err := jsonquery.New(node).ArrayRequired(responseKey)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		cursor, err := jsonquery.New(node, "response_metadata").StringOptional("next_cursor")
		if err != nil {
			return "", err
		}

		if cursor == nil || *cursor == "" {
			return "", nil
		}

		return *cursor, nil
	}
}
