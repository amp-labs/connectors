package pinterest

import (
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const pageSize = 250

func nextRecordsURL(reqLink *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).StringOptional("bookmark")
		if err != nil {
			return "", err
		}

		if pagination != nil {
			nextURL := *reqLink

			query := nextURL.Query()

			query.Set("bookmark", *pagination)
			query.Set("page_size", strconv.Itoa(pageSize))

			nextURL.RawQuery = query.Encode()

			return nextURL.String(), nil
		}

		return "", nil
	}
}
