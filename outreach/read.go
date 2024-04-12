package outreach

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		res    *common.JSONHTTPResponse
		err    error
		fields []string // Added this to satify the ParseResult function in Line 48
	)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// The NextPage URL has all the necessary parameters.
		res, err = c.get(ctx, config.NextPage)
		if err != nil {
			return nil, err
		}
	} else {
		q := makeQueryValues(config)
		fullURL, err := url.JoinPath(c.BaseURL, config.ObjectName)
		if err != nil {
			return nil, err
		}

		fullURL += q

		res, err = c.get(ctx, fullURL)
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(res, getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshaledData,
		fields,
	)
}

func makeQueryValues(config common.ReadParams) string {
	query := "?"

	// pageSize default values differ depending on endpoints
	if config.PageSize != 0 {
		query += fmt.Sprintf("page[size]=%v&", config.PageSize)
	}

	if len(config.FilterBy.Key) > 0 {
		switch deduceFilteringType(config) {
		case true:
			query += parseFilteringAttributes(config)
		default:
			query += fmt.Sprintf("filter[%s]=%v&", config.FilterBy.Key, config.FilterBy.Value)
		}
	}

	s := parseSortingQuery(config)
	query += s

	return query
}

// deduceFilteringType checks if the given parameters are
// slice of any data type. This helps to concatenate the values before
// sending the request.
func deduceFilteringType(config common.ReadParams) bool {
	_, OK := config.FilterBy.Value.([]any)

	return OK
}

func parseFilteringAttributes(config common.ReadParams) string {
	var query string

	if len(config.FilterBy.Key) > 0 {
		filter := config.FilterBy.Key

		// Check the data type of the provided value
		// is string. If it's create the filter query
		if arr, ok := config.FilterBy.Value.([]string); ok {
			values := strings.Join(arr, ",")
			query = fmt.Sprintf("filter[%s]=%v", filter, values)
		}

		// Checking if the type of the values is integer
		// If it's create the filter query
		if arr, ok := config.FilterBy.Value.([]int); ok {
			values := make([]string, len(arr))
			for i, v := range arr {
				values[i] = strconv.Itoa(v)
			}

			nv := strings.Join(values, ",")
			query = fmt.Sprintf("filter[%s]=%v&", filter, nv)
		}
	}

	return query
}

func parseSortingQuery(config common.ReadParams) string {
	var query string

	sorter := config.SortBy.Key
	if len(sorter) > 0 {
		switch config.SortBy.Value {
		case common.Ascending:
			query = fmt.Sprintf("sort=%s&", sorter)
		case common.Descending:
			query = fmt.Sprintf("sort=-%s&", sorter)
		}
	}

	return query
}
