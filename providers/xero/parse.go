package xero

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextCursor, err := jsonquery.New(node).StringOptional("next_cursor")
		if err != nil || nextCursor == nil {
			return "", err
		}

		return *nextCursor, nil
	}
}
