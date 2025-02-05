package stripe

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// Pagination is implemented as follows:
//   - Check the response to determine if there are more items to retrieve.
//   - If additional items exist, extract the ID of the last item from the current page.
//   - Use this ID to query the next page, starting after the last item ID from the current page.
//
// For more details, refer to the documentation:
// https://docs.stripe.com/api/pagination?lang=curl
func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasMore, err := jsonquery.New(node).BoolWithDefault("has_more", false)
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil
		}

		data, err := jsonquery.New(node).Array("data", true)
		if err != nil {
			return "", err
		}

		if len(data) == 0 {
			return "", nil
		}

		lastElement := data[len(data)-1]

		lastItemID, err := jsonquery.New(lastElement).Str("id", true)
		if err != nil {
			return "", err
		}

		if lastItemID == nil {
			return "", nil
		}

		url.WithQueryParam("starting_after", *lastItemID)

		return url.String(), nil
	}
}
