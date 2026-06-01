package monday

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func marshalItemsReadResult(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	for i := range records {
		flattenItemColumnValues(records[i])
	}

	return common.GetMarshaledData(records, fields)
}

func extractItemsRecords(node *ajson.Node) ([]map[string]any, error) {
	dataNode, err := node.GetKey("data")
	if err != nil {
		return nil, err
	}

	boards, err := jsonquery.New(dataNode).ArrayOptional("boards")
	if err != nil || len(boards) == 0 {
		return []map[string]any{}, nil
	}

	itemsPage, err := jsonquery.New(boards[0]).ObjectOptional("items_page")
	if err != nil || itemsPage == nil {
		return []map[string]any{}, nil
	}

	records, err := jsonquery.New(itemsPage).ArrayOptional("items")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

func makeItemsNextRecordsURL(limit int) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		dataNode, err := node.GetKey("data")
		if err != nil {
			return "", nil //nolint:nilerr
		}

		boards, err := jsonquery.New(dataNode).ArrayOptional("boards")
		if err != nil || len(boards) == 0 {
			return "", nil //nolint:nilerr
		}

		itemsPage, err := jsonquery.New(boards[0]).ObjectOptional("items_page")
		if err != nil || itemsPage == nil {
			return "", nil //nolint:nilerr
		}

		items, err := jsonquery.New(itemsPage).ArrayOptional("items")
		if err != nil || len(items) < limit {
			return "", nil //nolint:nilerr
		}

		cursor, err := jsonquery.New(itemsPage).TextWithDefault("cursor", "")
		if err != nil || cursor == "" {
			return "", nil //nolint:nilerr
		}

		return cursor, nil
	}
}
