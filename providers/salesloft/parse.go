package salesloft

import (
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

/*
Response example:

	{
	  "metadata": {
	    "filtering": {},
	    "paging": {
	      "per_page": 25,
	      "current_page": 1,
	      "next_page": 2,
	      "prev_page": null
	    },
	    "sorting": {
	      "sort_by": "updated_at",
	      "sort_direction": "DESC NULLS LAST"
	    }
	  },
	  "data": [...]
	}
*/

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).ArrayRequired("data")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// If the request URL has a sort_direction param, we're using cursor-based polling.
		// Extract the last record's updated_at and use it as the cursor for the next request.
		// This avoids deep pagination (page 500+) which causes rate limit cost escalation
		// and server errors on Salesloft's API.
		if _, hasSortDir := reqLink.GetFirstQueryParam("sort_direction"); hasSortDir {
			return makeCursorNextPageURL(reqLink, node)
		}

		// Fallback to offset-based pagination for objects not in the cursor-based allow-list
		// (e.g., users, tasks, meetings).
		return makeOffsetNextPageURL(reqLink, node)
	}
}

// makeCursorNextPageURL extracts the last record's updated_at timestamp from the response
// and builds the next request URL using it as a cursor (updated_at[gt]=<timestamp>).
// Returns empty string when no records are present, signaling that reading is complete.
func makeCursorNextPageURL(reqLink *urlbuilder.URL, node *ajson.Node) (string, error) {
	records, err := jsonquery.New(node).ArrayOptional("data")
	if err != nil || len(records) == 0 {
		return "", err
	}

	// If the number of records is less than the page size, this is the last page.
	// No need to fetch another page — we have all the data.
	if len(records) < DefaultPageSize {
		return "", nil
	}

	// Get the last record's updated_at value to use as the cursor.
	lastRecord := records[len(records)-1]

	updatedAt, err := jsonquery.New(lastRecord).StringOptional("updated_at")
	if err != nil {
		return "", err
	}

	if updatedAt == nil {
		// No updated_at on the last record, cannot advance cursor.
		return "", nil
	}

	// Build the next URL with the cursor. Use updated_at[gt] (strict greater than)
	// to avoid re-fetching the last record. Remove any page parameter and the
	// previous updated_at[gte] filter since we're advancing the cursor.
	reqLink.RemoveQueryParam("page")
	reqLink.RemoveQueryParam("updated_at[gte]")
	reqLink.WithQueryParam("updated_at[gt]", *updatedAt)

	return reqLink.String(), nil
}

// makeOffsetNextPageURL extracts the next page number from the response metadata
// and builds the next request URL using offset-based pagination (page=N).
func makeOffsetNextPageURL(reqLink *urlbuilder.URL, node *ajson.Node) (string, error) {
	nextPageNum, err := jsonquery.New(node, "metadata", "paging").IntegerOptional("next_page")
	if err != nil {
		if errors.Is(err, jsonquery.ErrKeyNotFound) {
			// list resource doesn't support pagination, hence no next page
			return "", nil
		}

		return "", err
	}

	if nextPageNum == nil {
		// next page doesn't exist
		return "", nil
	}

	// use request URL to infer the next page URL
	reqLink.WithQueryParam("page", strconv.FormatInt(*nextPageNum, 10))

	return reqLink.String(), nil
}
