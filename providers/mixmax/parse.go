package mixmax

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		switch objectName {
		case "appointmentlinks/me", "userpreferences/me", "users/me":
			data, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return nil, err
			}

			return []map[string]any{data}, nil
		default:
			return common.ExtractRecordsFromPath(responseField(objectName))(node)
		}
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasNext, err := jsonquery.New(node).BoolOptional("hasNext")
		if err != nil {
			return "", err
		}

		if (hasNext != nil) && *hasNext {
			next, err := jsonquery.New(node).StringOptional("next")
			if err != nil {
				return "", err
			}

			if next == nil {
				return "", nil
			}

			return *next, nil
		}

		return "", nil
	}
}
