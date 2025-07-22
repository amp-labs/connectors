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
	return func(node *ajson.Node) (string, error) {
		if url == nil {
			return "", nil
		}

		nextURL := *url

		// Try to parse pagination object from the response
		nextURLStr, shouldStop := handlePaginationObject(node, &nextURL)

		if nextURLStr != "" {
			return nextURLStr, nil
		}

		if shouldStop {
			return "", nil // No more pages, stop pagination
		}

		// Fallback: no pagination object, increment based on URL query
		return handleURLQueryFallback(&nextURL)
	}
}

// handlePaginationObject extracts pagination info from response.
// Returns: (nextURL, shouldStop).
// - nextURL: the next page URL if pagination is valid and not at end.
// - shouldStop: true if we've reached the last page.
func handlePaginationObject(n *ajson.Node, url *url.URL) (string, bool) {
	pagination, err := jsonquery.New(n).ObjectOptional("pagination")
	if err != nil || pagination == nil {
		return "", false
	}

	currentPage, err1 := jsonquery.New(pagination).IntegerOptional("currentPage")
	totalPages, err2 := jsonquery.New(pagination).IntegerOptional("pageCount")

	if err1 != nil || err2 != nil || currentPage == nil || totalPages == nil {
		return "", false
	}

	// Check if we've reached the last page
	if *currentPage >= *totalPages {
		return "", true
	}

	query := url.Query()
	query.Set(pageQuery, strconv.Itoa(int(*currentPage+1)))
	url.RawQuery = query.Encode()

	return url.String(), false
}

// handleURLQueryFallback handles pagination when no pagination object is present.
func handleURLQueryFallback(url *url.URL) (string, error) {
	query := url.Query()
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
	url.RawQuery = query.Encode()

	return url.String(), nil
}
