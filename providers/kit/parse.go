package kit

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).Object("pagination", true)
		if err != nil {
			return "", err
		}

		if pagination != nil {
			hasNextPage, err := jsonquery.New(pagination).BoolOptional("has_next_page")
			if err != nil {
				return "", err
			}

			if !(*hasNextPage) {
				return "", nil
			}

			endCursorToken, err := jsonquery.New(pagination).StringOptional("end_cursor")
			if err != nil {
				return "", err
			}

			reqLink.WithQueryParam("after", *endCursorToken)

			return reqLink.String(), nil
		}

		return "", nil
	}
}
