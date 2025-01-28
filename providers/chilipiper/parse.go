package chilipiper

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
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
func nextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		jsonQuery := jsonquery.New(node)

		page, err := jsonQuery.Integer(pageKey, false)
		if err != nil {
			return "", err
		}

		totalRecords, err := jsonQuery.Integer(totalKey, false)
		if err != nil {
			return "", err
		}

		pagesize, err := strconv.Atoi(pageSize)
		if err != nil {
			return "", err
		}

		if hasMorePages(pagesize, int(*page), int(*totalRecords)) {
			pg := strconv.Itoa(int(*page + 1))

			url.WithQueryParam(pageKey, pg)

			return url.String(), nil
		}

		return "", nil
	}
}

func hasMorePages(pageSize, page, total int) bool {
	if total < pageSize {
		return false
	}

	return !(total <= ((page + 1) * pageSize))
}
