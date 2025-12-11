package chilipiper

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Response structure
//
// {
// 	"results": [...],
// 	"total": 0,
// 	"page": 0,
// 	"pageSize": 0
// }

// nextRecordsURL builds the next-page url func.
// https://developer.close.com/topics/pagination/
func nextRecordsURL(url string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		jsonQuery := jsonquery.New(node)

		page, err := jsonQuery.IntegerRequired(pageKey)
		if err != nil {
			return "", err
		}

		totalRecords, err := jsonQuery.IntegerRequired(totalKey)
		if err != nil {
			return "", err
		}

		pagesize, err := strconv.Atoi(readpageSize)
		if err != nil {
			return "", err
		}

		if hasMorePages(pagesize, int(page), int(totalRecords)) {
			pg := strconv.Itoa(int(page + 1))

			nextURL, err := urlbuilder.New(url)
			if err != nil {
				return "", err
			}

			nextURL.WithQueryParam(pageKey, pg)

			return nextURL.String(), nil
		}

		return "", nil
	}
}

func hasMorePages(pageSize, page, total int) bool {
	if total < pageSize {
		return false
	}

	return !(total <= ((page + 1) * pageSize)) // nolint:staticcheck
}
