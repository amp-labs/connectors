package clari

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		switch objectName {
		case adminLimit:
			data, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return nil, err
			}

			return []map[string]any{data}, nil
		default:
			return common.GetRecordsUnderJSONPath(responseField(objectName))(node)
		}
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextURL, err := jsonquery.New(node).StringOptional("nextLink")
		if err != nil {
			return "", err
		}

		if nextURL == nil {
			return "", nil
		}

		return *nextURL, nil
	}
}
