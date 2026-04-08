package supersend

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/supersend/metadata"
	"github.com/spyzhov/ajson"
)

// Pagination constants for SuperSend API.
// SuperSend uses offset-based pagination with limit/offset query parameters.
// The API returns pagination.has_more to indicate if more records exist.
const (
	defaultPageSize = "100" // Default page size for SuperSend API (max is 100)
	limitParam      = "limit"
	offsetParam     = "offset"

	// updatedAtField is the timestamp field used for connector-side filtering.
	// SuperSend API doesn't support native time-based filtering, so we filter
	// records client-side using the updatedAt field for incremental sync.
	// Format: ISO 8601 / RFC3339 (e.g., "2024-01-15T10:00:00.000Z").
	updatedAtField = "updatedAt"
)

// buildReadRequest constructs the HTTP request for read operations.
// Handles pagination via offset parameter and respects PageSize up to max limit.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		// Use NextPage URL directly for pagination
		nextPageURL, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, nextPageURL.String(), nil)
	}

	// Build initial URL from metadata
	apiURL, err := c.buildURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Add pagination limit - use PageSize if provided, otherwise use default.
	// SuperSend API enforces max of 100 on its side.
	apiURL.WithQueryParam(limitParam, readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	return http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
}

func (c *Connector) buildURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// LookupArrayFieldName returns the responseKey from the schema
	responseKey := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, params.ObjectName)
	nextPageFunc := makeNextRecordsURL(request.URL)

	return common.ParseResultFiltered(
		params,
		response,
		getRecords(responseKey),
		makeFilterFunc(params, nextPageFunc),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

// makeFilterFunc returns a filter function for connector-side time-based filtering.
// SuperSend API doesn't support native time filtering, so we filter records
// client-side using the updatedAt field when Since/Until params are provided.
func makeFilterFunc(params common.ReadParams, nextPageFunc common.NextPageFunc) common.RecordsFilterFunc {
	// If no time filtering is requested, use identity filter (no filtering)
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	// Apply time-based filtering using updatedAt field.
	// Using Unordered since SuperSend doesn't guarantee record ordering.
	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		updatedAtField,
		time.RFC3339,
		nextPageFunc,
	)
}

// getRecords returns a function that extracts records from the response.
// Uses slices for nested paths (e.g., "data.conversations" is split into
// nestedPath=["data"] and jsonPath="conversations").
func getRecords(responseKey string) common.NodeRecordsFunc {
	parts := strings.Split(responseKey, ".")
	if len(parts) > 1 {
		return common.MakeRecordsFunc(parts[len(parts)-1], parts[:len(parts)-1]...)
	}

	return common.MakeRecordsFunc(responseKey)
}

// makeNextRecordsURL returns a function that builds the next page URL if more records exist.
// SuperSend uses pagination.has_more to indicate if there are more records.
func makeNextRecordsURL(requestURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if !hasMoreRecords(node) {
			return "", nil
		}

		// Calculate next offset based on current request
		nextOffset := calculateNextOffset(requestURL)

		return buildNextPageURL(requestURL, nextOffset)
	}
}

// hasMoreRecords checks the pagination.has_more field to determine if more records exist.
func hasMoreRecords(node *ajson.Node) bool {
	hasMore, _ := jsonquery.New(node, "pagination").BoolWithDefault("has_more", false)

	return hasMore
}

// calculateNextOffset extracts current offset from URL and adds the limit to get next offset.
func calculateNextOffset(requestURL *url.URL) int {
	query := requestURL.Query()

	currentOffset := 0

	if offsetStr := query.Get(offsetParam); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil {
			currentOffset = parsed
		}
	}

	// Get limit from URL, default to defaultPageSize if not present or invalid
	limit, _ := strconv.Atoi(defaultPageSize)

	if limitStr := query.Get(limitParam); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	return currentOffset + limit
}

// buildNextPageURL constructs the URL for the next page of results.
func buildNextPageURL(requestURL *url.URL, nextOffset int) (string, error) {
	nextURL, err := urlbuilder.FromRawURL(requestURL)
	if err != nil {
		return "", err
	}

	nextURL.WithQueryParam(offsetParam, strconv.Itoa(nextOffset))

	return nextURL.String(), nil
}
