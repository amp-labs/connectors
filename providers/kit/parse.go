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
			hasNextPage, err := jsonquery.New(pagination).Bool("has_next_page", true)
			if err != nil {
				return "", err
			}

			if !(*hasNextPage) {
				return "", nil
			}

			endCursorToken, err := jsonquery.New(pagination).Str("end_cursor", true)
			if err != nil {
				return "", err
			}

			reqLink.WithQueryParam("after", *endCursorToken)

			// Next page token is base64 encoded,
			// therefore it may contain symbols that should be exempt from escaping.
			reqLink.AddEncodingExceptions(map[string]string{
				"%3D": "=",
			})

			return reqLink.String(), nil
		}

		return "", nil
	}
}
