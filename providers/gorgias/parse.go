package gorgias

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		switch objectName {
		case account:
			data, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return nil, err
			}

			return []map[string]any{data}, nil
		default:
			return common.ExtractRecordsFromPath(dataField)(node)
		}
	}
}

func nextRecordsURL(root *ajson.Node) (string, error) {
	meta, err := jsonquery.New(root).ObjectOptional("meta")
	if err != nil {
		return "", err
	}

	nextCursor, err := jsonquery.New(meta).StringOptional("next_cursor")
	if err != nil {
		return "", err
	}

	if nextCursor == nil {
		return "", nil
	}

	return *nextCursor, nil
}
