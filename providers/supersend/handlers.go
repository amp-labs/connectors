package supersend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
// See: https://documenter.getpostman.com/view/19579115/2sA3kSo3FD
const (
	defaultPageSize = "100" // Default page size for SuperSend API (max is 100)
	limitParam      = "limit"
	offsetParam     = "offset"

	// updatedAtField is the timestamp field used for connector-side filtering.
	// SuperSend API doesn't support native time-based filtering, so we filter
	// records client-side using the updatedAt field for incremental sync.
	// Format: ISO 8601 / RFC3339 (e.g., "2024-01-15T10:00:00.000Z").
	// See: https://documenter.getpostman.com/view/19579115/2sA3kSo3FD
	updatedAtField = "updatedAt"
)

// writePathConfig defines the write/delete path configuration for each object.
// SuperSend API uses different paths for read vs write/delete operations.
// See: https://documenter.getpostman.com/view/19579115/2sA3kSo3FD
type writePathConfig struct {
	createPath string // Path for POST (create) - without record ID
	updatePath string // Path for PUT (update) - record ID will be appended
	deletePath string // Path for DELETE - record ID will be appended
	usesPatch  bool   // Whether to use PATCH instead of PUT for updates
}

// objectWritePaths maps object names to their write/delete path configurations.
// nolint:gochecknoglobals
var objectWritePaths = map[string]writePathConfig{
	"labels": {
		createPath: "/v1/labels",
		updatePath: "/v1/labels",
		deletePath: "/v1/labels",
	},
	"senders": {
		createPath: "/v1/sender",
		updatePath: "/v1/sender",
		deletePath: "", // No delete endpoint for senders
	},
	"teams": {
		createPath: "/v2/teams",
		updatePath: "", // No update endpoint documented
		deletePath: "", // No delete endpoint documented
	},
	"campaigns": {
		createPath: "/v1/auto/campaign",
		updatePath: "/v1/campaign",
		deletePath: "/v1/auto/campaign",
	},
	"contacts": {
		createPath: "/v2/contacts",
		updatePath: "/v2/contacts",
		deletePath: "/v2/contacts",
		usesPatch:  true, // V2 API uses PATCH for updates
	},
	"sender-profiles": {
		createPath: "/v1/sender-profile",
		updatePath: "/v1/sender-profile",
		deletePath: "/v1/sender-profile",
	},
}

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

// buildWriteRequest constructs the HTTP request for write operations.
// Uses POST for create (no RecordId) and PUT/PATCH for update (with RecordId).
// SuperSend API uses different paths for write vs read operations.
// See: https://documenter.getpostman.com/view/19579115/2sA3kSo3FD
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	config, ok := objectWritePaths[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	apiURL, method, err := c.buildWriteURL(params, config)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, apiURL.String(), bytes.NewReader(jsonData))
}

// buildWriteURL constructs the URL and HTTP method for write operations.
func (c *Connector) buildWriteURL(params common.WriteParams, config writePathConfig) (*urlbuilder.URL, string, error) {
	if params.IsUpdate() {
		return c.buildUpdateURL(params, config)
	}

	return c.buildCreateURL(config)
}

// buildUpdateURL constructs the URL and method for update operations.
func (c *Connector) buildUpdateURL(params common.WriteParams, config writePathConfig) (*urlbuilder.URL, string, error) {
	if config.updatePath == "" {
		return nil, "", common.ErrOperationNotSupportedForObject
	}

	apiURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, config.updatePath, params.RecordId)
	if err != nil {
		return nil, "", err
	}

	method := http.MethodPut
	if config.usesPatch {
		method = http.MethodPatch
	}

	return apiURL, method, nil
}

// buildCreateURL constructs the URL and method for create operations.
func (c *Connector) buildCreateURL(config writePathConfig) (*urlbuilder.URL, string, error) {
	if config.createPath == "" {
		return nil, "", common.ErrOperationNotSupportedForObject
	}

	apiURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, config.createPath)
	if err != nil {
		return nil, "", err
	}

	return apiURL, http.MethodPost, nil
}

// parseWriteResponse parses the response from write operations.
// SuperSend returns responses wrapped in a "data" key with a "success" field.
func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// SuperSend wraps responses in "data" key
	dataNode, err := jsonquery.New(body).ObjectOptional("data")
	if err != nil {
		return nil, err
	}

	// Use the data node if available, otherwise use the body directly
	responseNode := body
	if dataNode != nil {
		responseNode = dataNode
	}

	// Extract record ID from response
	recordID, err := jsonquery.New(responseNode).StringOptional("id")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(responseNode)
	if err != nil {
		return nil, err
	}

	result := &common.WriteResult{
		Success: true,
		Data:    data,
	}

	if recordID != nil {
		result.RecordId = *recordID
	}

	return result, nil
}

// buildDeleteRequest constructs the HTTP request for delete operations.
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	config, ok := objectWritePaths[params.ObjectName]
	if !ok || config.deletePath == "" {
		return nil, common.ErrOperationNotSupportedForObject
	}

	apiURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, config.deletePath, params.RecordId)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, apiURL.String(), nil)
}

// parseDeleteResponse parses the response from delete operations.
func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	// Validate HTTP status code for delete operations
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
