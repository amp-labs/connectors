package recurly

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasMore, err := jsonquery.New(node).BoolRequired("has_more")
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil //nolint: nilerr
		}

		next, err := jsonquery.New(node).StringRequired("next")
		if err != nil {
			return "", err
		}

		return next, nil
	}
}
