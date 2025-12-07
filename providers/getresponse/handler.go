package getresponse

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "perPage"
	pageSize    = "100"
	pageKey     = "page"
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
	url.WithQueryParam(pageSizeKey, pageSize)
	url.WithQueryParam(pageKey, "1")

	// Add field selection if specified
	if len(params.Fields.List()) > 0 {
		url.WithQueryParam("fields", strings.Join(params.Fields.List(), ","))
	}

	// Parse GetResponse-specific filter and sort from params.Filter
	// Format: "query[name]=value&query[isDefault]=true&sort[name]=ASC&sort[createdOn]=DESC"
	// This is a simple implementation - can be extended for more complex filtering
	if params.Filter != "" {
		addGetResponseFilters(url, params.Filter)
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
// Multiple filters can be separated by &, e.g., "query[name]=test&sort[createdOn]=DESC"
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
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Add as unencoded query parameter (GetResponse uses bracket notation like query[name])
		// Brackets must remain unencoded: query[name] not query%5Bname%5D
		url.WithUnencodedQueryParam(key, value)
	}
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// GetResponse returns arrays directly, not wrapped in an object
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(""),
		makeNextRecordsURL(c, params, request),
		common.GetMarshaledData,
		params.Fields,
	)
}

// makeNextRecordsURL constructs the next page URL based on GetResponse pagination.
// GetResponse uses response headers (TotalCount, TotalPages, CurrentPage) for pagination info,
// but since we only have the response body here, we check if the current page has records.
// If the response is empty, we're done. Otherwise, increment the page.
func makeNextRecordsURL(c *Connector, params common.ReadParams, request *http.Request) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Check if response has any records - if empty array, we're done
		records, err := jsonquery.New(node).ArrayOptional("")
		if err != nil || len(records) == 0 {
			return "", nil //nolint:nilerr
		}

		// Extract current page from request URL
		currentPageStr := request.URL.Query().Get(pageKey)
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

		// Rebuild URL with incremented page
		path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
		if err != nil {
			return "", nil //nolint:nilerr
		}

		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
		if err != nil {
			return "", nil //nolint:nilerr
		}

		url.WithQueryParam(pageSizeKey, pageSize)
		url.WithQueryParam(pageKey, strconv.Itoa(nextPage))

		return url.String(), nil
	}
}
