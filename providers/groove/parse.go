package groove

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node, "meta").ObjectOptional("pagination")
		if err != nil {
			return "", err
		}

		nextPage, err := jsonquery.New(pagination).StringOptional("next_page")
		if err != nil {
			return "", err
		}

		// Check this to avoid panics
		if nextPage == nil {
			return "", nil
		}

		return *nextPage, nil
	}
}
