package clickup

import (
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(previousURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		lastPage, err := jsonquery.New(node).BoolOptional("last_page")
		if err != nil {
			return "", err
		}

		if lastPage == nil || *lastPage {
			return "", nil
		}

		nextURL := *previousURL

		// parse the current page value form previous request url
		query := nextURL.Query()
		currentPage := 0
		currentPageStr := query.Get("page")

		if currentPageStr != "" {
			currentPage, err = strconv.Atoi(currentPageStr)
			if err != nil {
				return "", nil //nolint:nilerr
			}
		}

		nextPage := currentPage + 1

		query.Set("page", strconv.Itoa(nextPage))
		nextURL.RawQuery = query.Encode()

		return nextURL.String(), nil
	}
}
