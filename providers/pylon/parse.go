package pylon

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).ObjectOptional("pagination")
		if err != nil || pagination == nil {
			return "", err
		}

		hasNextPage, err := jsonquery.New(pagination).BoolRequired("has_next_page")
		if err != nil || !hasNextPage {
			return "", nil //nolint:nilerr
		}

		nextCursor, err := jsonquery.New(pagination).StringOptional("cursor")
		if err != nil || nextCursor == nil {
			return "", err
		}

		return *nextCursor, nil
	}
}
