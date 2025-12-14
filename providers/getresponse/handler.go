package getresponse

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Pagination constants for GetResponse APIv3.
	// According to GetResponse API documentation (https://apidocs.getresponse.com/v3):
	// - Default page size: 100 resources per page
	// - Maximum page size: 1000 resources per page
	// - Minimum page size: 1 resource per page
	// The maximum value (1000) is used to minimize the number of API calls when iterating over all pages.
	// This page size works for all endpoints including: contacts, campaigns, forms, custom-events, etc.
	pageSizeKey = "perPage"
	pageSize    = "1000" // Maximum allowed by GetResponse API to minimize API calls
	pageKey     = "page"
	sinceKey    = "query[createdOn][from]"
	untilKey    = "query[createdOn][to]"
	apiVersion  = "v3"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	// Set pagination parameters
	// Use maximum page size (1000) if not specified or if it exceeds the maximum
	requestedPageSize := params.PageSize
	maxPageSizeInt := 1000 // Maximum allowed by GetResponse API
	if requestedPageSize <= 0 || requestedPageSize > maxPageSizeInt {
		requestedPageSize = maxPageSizeInt
	}
	url.WithQueryParam(pageSizeKey, strconv.Itoa(requestedPageSize))
	url.WithQueryParam(pageKey, "1")

	// Add field selection
	url.WithQueryParam("fields", strings.Join(params.Fields.List(), ","))

	// Parse GetResponse-specific filter and sort from params.Filter
	// Format: "query[name]=value&query[isDefault]=true&sort[name]=ASC&sort[createdOn]=DESC"
	// This is a simple implementation - can be extended for more complex filtering
	if params.Filter != "" {
		addGetResponseFilters(url, params.Filter)
	}

	// Only add provider-side since/until filters if the object supports them
	if shouldAddProviderSideFilter(params.ObjectName, params) {
		if !params.Since.IsZero() {
			url.WithQueryParam(sinceKey, datautils.Time.FormatRFC3339inUTC(params.Since))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam(untilKey, datautils.Time.FormatRFC3339inUTC(params.Until))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// addGetResponseFilters parses GetResponse filter string and adds query/sort parameters.
// Format examples:
//   - "query[name]=campaign_name" -> adds query[name]=campaign_name
//   - "query[isDefault]=true" -> adds query[isDefault]=true
//   - "sort[name]=ASC" -> adds sort[name]=ASC
//   - "sort[createdOn]=DESC" -> adds sort[createdOn]=DESC
//
// Multiple filters can be separated by &, e.g., "query[name]=test&sort[createdOn]=DESC".
func addGetResponseFilters(url *urlbuilder.URL, filterStr string) {
	// Simple parser for GetResponse filter format
	// Split by & to get individual filter clauses
	filters := strings.Split(filterStr, "&")
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(filter, "=", 2) // nolint:mnd
		if len(parts) != 2 {                    // nolint:mnd
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		url.WithQueryParam(key, value)
	}
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// GetResponse returns arrays directly, not wrapped in an object
	// Use ParseResultFiltered to support connector-side filtering for objects that don't support provider-side filtering
	return common.ParseResultFiltered(
		params,
		response,
		common.MakeRecordsFunc(""),
		makeFilterFunc(params, request.URL),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

// makeNextRecordsURL constructs the next page URL based on GetResponse pagination.
// GetResponse uses response headers (TotalCount, TotalPages, CurrentPage) for pagination info,
// but since we only have the response body here, we check if the current page has records.
// If the response is empty, we're done. Otherwise, increment the page.
func makeNextRecordsURL(requestURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Check if response has any records - if empty array, we're done
		records, err := jsonquery.New(node).ArrayOptional("")
		if err != nil || len(records) == 0 {
			return "", nil //nolint:nilerr
		}

		// Extract current page from request URL
		currentPageStr := requestURL.Query().Get(pageKey)
		if currentPageStr == "" {
			currentPageStr = "1"
		}

		currentPage, err := strconv.Atoi(currentPageStr)
		if err != nil {
			return "", nil //nolint:nilerr
		}

		// Parse pageSize to int for comparison
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil {
			return "", nil //nolint:nilerr
		}

		// If we got fewer records than pageSize, this is the last page
		if len(records) < pageSizeInt {
			return "", nil
		}

		// Increment page for next request
		nextPage := currentPage + 1

		url, err := urlbuilder.FromRawURL(requestURL)
		if err != nil {
			return "", nil //nolint:nilerr
		}

		// Update only the page number - pageSize is already set in the original request
		url.WithQueryParam(pageKey, strconv.Itoa(nextPage))

		return url.String(), nil
	}
}
