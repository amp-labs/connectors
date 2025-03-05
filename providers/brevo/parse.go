package brevo

import (
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "limit"
	pageSize    = 200
)

func nextRecordsURL(previousURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		totalCount, err := jsonquery.New(node).IntegerOptional("count")
		if err != nil || totalCount == nil {
			return "", err
		}

		nextURL := *previousURL

		// parse the current offset value form previous request url
		query := nextURL.Query()
		currentOffset := 0
		offsetStr := query.Get("offset")

		if offsetStr != "" {
			currentOffset, err = strconv.Atoi(offsetStr)
			if err != nil {
				return "", nil //nolint:nilerr
			}
		}

		// calcuate the next offset value
		nextOffset := currentOffset + pageSize

		// if nextOffset would exceed total count, we've fetched all records
		if nextOffset >= int(*totalCount) {
			return "", nil
		}

		query.Set("offset", strconv.Itoa(nextOffset))
		query.Set("limit", strconv.Itoa(pageSize))
		nextURL.RawQuery = query.Encode()

		return nextURL.String(), nil
	}
}
