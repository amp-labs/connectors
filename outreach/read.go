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

		// Testing pagination
		fullURL += q
		fmt.Println("The fullURL: ", fullURL)

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
	var q = "?"

	// pageSize default values differ depending on endpoints
	if config.PageSize != 0 {
		q += fmt.Sprintf("page[size]=%v&", config.PageSize)
	}

	if len(config.FilterBy.Key) > 0 {
		switch deduceFilteringType(config) {
		case true:
			q = q + parseFilteringAttributes(config)
		default:
			q += fmt.Sprintf("filter[%s]=%v&", config.FilterBy.Key, config.FilterBy.Value)
		}
	}

	s := parseSortingQuery(config)
	q += s

	fmt.Println(q)

	return q
}

// deduceFilteringType checks if the given parameters are
// slice of any data type. This helps to concatenate the values before
// sending the request.
func deduceFilteringType(config common.ReadParams) bool {
	_, OK := config.FilterBy.Value.([]any)
	return OK
}

func parseFilteringAttributes(config common.ReadParams) string {
	var q string

	if len(config.FilterBy.Key) > 0 {
		filter := config.FilterBy.Key

		// Check the data type of the provided value
		// is string. If it's create the filter query
		fmt.Println("Values")
		if arr, ok := config.FilterBy.Value.([]string); ok {
			values := strings.Join(arr, ",")
			q = fmt.Sprintf("filter[%s]=%v", filter, values)
		}

		// Checking if the type of the values is integer
		// If it's create the filter query
		if arr, ok := config.FilterBy.Value.([]int); ok {
			values := make([]string, len(arr))
			for i, v := range arr {
				values[i] = strconv.Itoa(v)
			}

			nv := strings.Join(values, ",")
			q = fmt.Sprintf("filter[%s]=%v&", filter, nv)
		}
	}

	return q
}

func parseSortingQuery(config common.ReadParams) string {
	var q string
	sorter := config.SortBy.Key

	if len(sorter) > 0 {
		switch config.SortBy.Value {
		case 1:
			q = fmt.Sprintf("sort=-%s", sorter)
		default:
			q = fmt.Sprintf("sort=%s", sorter)
		}
	}

	return q
}
