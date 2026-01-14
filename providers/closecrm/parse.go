package closecrm

import (
	"errors"

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

		if hasMore != nil {
			if *hasMore {
				currSkip, exists := url.GetFirstQueryParam(skipQuery)
				if !exists {
					return "", ErrSkipFailure
				}

				url.WithQueryParam(skipQuery, currSkip+defaultPageSize)
				url.WithQueryParam(limitQuery, defaultPageSize)

				return addTrailingSlashIfNeeded(url.String()), nil
			}
		}

		return "", nil
	}
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
