package asana

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPageURL, err := jsonquery.New(node, "next_page").StringOptional("uri")
		if err != nil {
			return "", err
		}

		if nextPageURL == nil {
			return "", nil
		}

		return *nextPageURL, nil
	}
}
