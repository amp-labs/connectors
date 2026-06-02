package monday

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func extractItemsRecords(node *ajson.Node) ([]*ajson.Node, error) {
	dataNode, err := node.GetKey("data")
	if err != nil {
		return nil, err
	}

	boards, err := jsonquery.New(dataNode).ArrayOptional("boards")
	if err != nil {
		return nil, err
	}

	if len(boards) == 0 {
		return []*ajson.Node{}, nil
	}

	itemsPage, err := jsonquery.New(boards[0]).ObjectOptional("items_page")
	if err != nil {
		return nil, err
	}

	if itemsPage == nil {
		return []*ajson.Node{}, nil
	}

	return jsonquery.New(itemsPage).ArrayOptional("items")
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
