package calendly

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	dataKey = "collection"
)

func nextRecordsURL(root *ajson.Node) (string, error) {
	pagination, err := jsonquery.New(root).ObjectOptional("pagination")
	if err != nil {
		return "", err
	}

	nextCursor, err := jsonquery.New(pagination).StringOptional("next_page")
	if err != nil {
		return "", err
	}

	if nextCursor == nil {
		return "", nil
	}

	return *nextCursor, nil
}
