package paddle

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasMore, err := jsonquery.New(node, "meta", "pagination").BoolWithDefault("has_more", false)
		if err != nil || !hasMore {
			return "", nil //nolint: nilerr
		}

		next, err := jsonquery.New(node, "meta", "pagination").StringOptional("next")
		if err != nil || next == nil || *next == "" {
			return "", nil //nolint:nilerr
		}

		return *next, nil
	}
}
