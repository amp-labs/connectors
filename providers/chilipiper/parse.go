package chilipiper

import (
	"fmt"
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
func nextRecordsURL(url *urlbuilder.URL, objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		jsonQuery := jsonquery.New(node)

		if objectName == meetings {
			return constructNextPageMeetings(url, jsonQuery)
		}

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

			nextURL, err := urlbuilder.New(url.String())
			if err != nil {
				return "", err
			}

			nextURL.WithQueryParam(pageKey, pg)

			return nextURL.String(), nil
		}

		return "", nil
	}
}

func constructNextPageMeetings(url *urlbuilder.URL, query *jsonquery.Query) (string, error) {
	hasMore, err := query.StringOptional("hasMore")
	if err != nil {
		return "", err
	}

	if *hasMore == "No" {
		return "", nil
	}

	value, exists := url.GetFirstQueryParam("page")
	if !exists {
		value = "0"
	}

	page, err := strconv.Atoi(value)
	if err != nil {
		return "", fmt.Errorf("constructing nextpage url : %w", err)
	}

	nextPage := page + 1

	url.WithQueryParam("page", strconv.Itoa(nextPage))

	return url.String(), nil
}

func hasMorePages(pageSize, page, total int) bool {
	if total < pageSize {
		return false
	}

	return !(total <= ((page + 1) * pageSize)) // nolint:staticcheck
}

func extractRecords(objectName string) common.RecordsFunc {
	if objectName == meetings {
		return common.ExtractRecordsFromPath("list", "data")
	}

	return common.ExtractRecordsFromPath("results")
}
