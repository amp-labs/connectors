package flatfile

import (
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL(url *url.URL) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {

		if url == nil {
			return "", nil
		}

		nextURL := *url

		query := nextURL.Query()
		currentPage := 1
		currentPageStr := query.Get(pageQuery)

		if currentPageStr != "" {
			var err error

			currentPage, err = strconv.Atoi(currentPageStr)
			if err != nil {
				return "", err
			}
		}

		nextPage := currentPage + 1
		query.Set(pageQuery, strconv.Itoa(nextPage))
		nextURL.RawQuery = query.Encode()

		return nextURL.String(), nil
	}
}
