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

// nolint: cyclop
func nextRecordsURL(url *url.URL) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		if url == nil {
			return "", nil
		}

		nextURL := *url

		// Try to parse pagination object from the response
		pagination, err := jsonquery.New(n).ObjectOptional("pagination")
		if err == nil && pagination != nil {
			currentPage, err1 := jsonquery.New(pagination).IntegerOptional("currentPage")
			totalPages, err2 := jsonquery.New(pagination).IntegerOptional("pageCount")

			if err1 == nil && err2 == nil && currentPage != nil && totalPages != nil {
				if *currentPage >= *totalPages {
					return "", nil
				}
				// If pagination is present, build the next page URL
				query := nextURL.Query()
				query.Set(pageQuery, strconv.Itoa(int(*currentPage+1)))
				nextURL.RawQuery = query.Encode()

				return nextURL.String(), nil
			}
		}

		// Fallback: no pagination object, increment based on URL query
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
