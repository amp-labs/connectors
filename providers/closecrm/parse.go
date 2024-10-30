package closecrm

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

/*
Response Schema:
{
    "has_more": false,
    "total_results": 1,
    "data": [
        {...},
		{...}
    ]
}

*/

// nextRecordsURL builds the next-page url func.
func nextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// check if there is more items in the collection.
		hasMore, err := jsonquery.New(node).Bool("has_more", false)
		if err != nil {
			return "", err
		}

		if *hasMore {
			url.WithQueryParam("start", strconv.FormatInt(*startValue, 10))

			return url.String(), nil
		}

		return "", nil
	}
}
