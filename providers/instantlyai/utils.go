package instantlyai

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const DefaultPageSize = 100

func makeNextRecordsURL(reqLink *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).StringRequired("next_starting_after")
		if err != nil {
			return "", err
		}

		if pagination != "" {
			nextLink := *reqLink
			query := nextLink.Query()
			query.Set("starting_after", pagination)
			nextLink.RawQuery = query.Encode()

			return nextLink.String(), nil
		}

		return "", nil
	}
}
