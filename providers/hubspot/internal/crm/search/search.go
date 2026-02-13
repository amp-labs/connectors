package search

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

const (
	// searchResultsLimit is the maximum number of records that can be returned by the search endpoint.
	// HubSpot's search API returns a 400 error if you try to paginate beyond this limit.
	// See: https://developers.hubspot.com/docs/api/crm/search#limitations
	searchResultsLimit = 10000

	// searchPageSize is the page size for search API requests.
	// Search endpoints support up to 200 records per page (unlike read endpoints which max at 100).
	searchPageSize = int64(200)
)

func (s Strategy) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	// Check if the NextPage token exceeds the search results limit.
	// HubSpot's search API returns a 400 error if you try to paginate beyond 10,000 records.
	// By detecting this proactively, we can return a specific error that callers can handle.
	if err := checkSearchResultsLimit(params.NextPage); err != nil {
		return nil, fmt.Errorf(
			"%w: requested offset %s exceeds limit %d",
			common.ErrResultsLimitExceeded,
			params.NextPage,
			searchResultsLimit,
		)
	}

	// Search has two execution paths:
	// - ObjectAPI: for core CRM objects supported by the canonical ObjectAPI endpoint.
	// - Non-ObjectAPI: for CRM objects not supported by ObjectAPI, which use separate endpoints (e.g., Lists).
	if core.ObjectsWithoutPropertiesAPISupport.Has(params.ObjectName) {
		return s.searchViaNonstandardSearchAPI(ctx, params)
	}

	return s.searchViaObjectAPI(ctx, params)
}

// checkSearchResultsLimit checks if the NextPage token exceeds HubSpot's search results limit.
// HubSpot's search API returns a 400 error if you try to paginate beyond 10,000 records.
// Returns ErrSearchResultsLimitExceeded if the limit is exceeded.
func checkSearchResultsLimit(nextPage common.NextPageToken) error {
	if nextPage == "" {
		return nil
	}

	// The NextPage token for HubSpot search is a numeric offset (as a string).
	offset, err := strconv.Atoi(string(nextPage))
	if err != nil {
		// If we can't parse the offset, it's not a numeric token (shouldn't happen for search).
		// Let the API call proceed and fail with a proper error if needed.
		return nil //nolint:nilerr
	}

	if offset >= searchResultsLimit {
		return common.ErrResultsLimitExceeded
	}

	return nil
}
