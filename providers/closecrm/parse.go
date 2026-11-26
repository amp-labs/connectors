package closecrm

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

/*
Read Response Schema:
{
    "has_more": false,
    "total_results": 1,
    "data": [{...},{...}]
}

*/

const (
	defaultPageSize = "100"      // nolint:gochecknoglobals
	maxPageSize     = 100        // nolint:gochecknoglobals
	limitQuery      = "_limit"   // nolint:gochecknoglobals
	skipQuery       = "_skip"    // nolint:gochecknoglobals
	hasMoreQuery    = "has_more" // nolint:gochecknoglobals
)

// ErrSkipFailure is an error generated when we fails to construct the next page url.
var ErrSkipFailure = errors.New("error: failed to create next page url")

// nextRecordsURL builds the next-page url func.
// https://developer.close.com/topics/pagination/
func nextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// check if there is more items in the collection.
		hasMore, err := jsonquery.New(node).BoolOptional(hasMoreQuery)
		if err != nil {
			return "", err
		}

		if hasMore == nil || !*hasMore {
			return "", nil
		}

		currSkip, exists := url.GetFirstQueryParam(skipQuery)
		if !exists {
			return "", ErrSkipFailure
		}

		// The page size is carried by the URL, so the next page keeps the same limit.
		currLimit, exists := url.GetFirstQueryParam(limitQuery)
		if !exists {
			currLimit = defaultPageSize
		}

		nextSkip, err := addPageSize(currSkip, currLimit)
		if err != nil {
			return "", err
		}

		url.WithQueryParam(skipQuery, nextSkip)
		url.WithQueryParam(limitQuery, currLimit)

		return addTrailingSlashIfNeeded(url.String()), nil
	}
}

// addPageSize returns the offset of the next page.
// Offsets are numbers sent as strings, so they are added as integers, not concatenated.
func addPageSize(currSkip, pageSize string) (string, error) {
	skip, err := strconv.Atoi(currSkip)
	if err != nil {
		return "", fmt.Errorf("%w: invalid %s value %q: %w", ErrSkipFailure, skipQuery, currSkip, err)
	}

	limit, err := strconv.Atoi(pageSize)
	if err != nil {
		return "", fmt.Errorf("%w: invalid %s value %q: %w", ErrSkipFailure, limitQuery, pageSize, err)
	}

	return strconv.Itoa(skip + limit), nil
}

/*
Search Response schema:

	{
		 "data": [{...},{...}],
		 "cursor": "..."
	}
*/
func getNextRecordCursor(node *ajson.Node) (string, error) {
	crs, err := jsonquery.New(node).StringOptional("cursor")
	if err != nil {
		return "", err
	}

	if crs == nil {
		return "", nil
	}

	return *crs, nil
}
