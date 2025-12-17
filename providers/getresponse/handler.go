package getresponse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
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
	pageSizeKey    = "perPage"
	pageSize       = "1000" // Maximum allowed by GetResponse API to minimize API calls
	maxPageSizeInt = 1000   // Maximum allowed by GetResponse API (int version for validation)
	pageKey        = "page"
	sinceKey       = "query[createdOn][from]"
	untilKey       = "query[createdOn][to]"
	apiVersion     = "v3"
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
	// Use maximum page size if not specified or if it exceeds the maximum
	requestedPageSize := params.PageSize
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
	for filter := range strings.SplitSeq(filterStr, "&") {
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

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

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

	// GetResponse API returns the object directly at the root level (not wrapped in "data")
	// According to swagger.json, 200 responses return the object schema directly
	// For 202 Accepted (create operations), there's no body, which is handled above

	// Extract record ID directly from ajson node using jsonquery.TextWithDefault
	// This avoids type conversion issues and handles string/int ID types automatically
	recordId := parseRecordIdFromNode(c, params.ObjectName, body)

	// Convert to map[string]any for WriteResult.Data (still needed for the result)
	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		// If conversion fails, return the error
		return nil, fmt.Errorf("failed to convert response body to map: %w", err)
	}

	return &common.WriteResult{Success: true, Data: data, RecordId: recordId}, nil
}

// parseRecordIdFromNode extracts the record ID directly from ajson.Node.
// This avoids type conversion issues with map[string]any and uses jsonquery.TextWithDefault
// to automatically handle string/int ID types.
//
// GetResponse uses object-specific ID fields (e.g., contactId, campaignId, addressId).
// This function uses the schemas (generated from swagger.json) to dynamically find the ID field.
//
// The function handles:
//   - Simple objects: "contacts" -> "contactId" (via schema lookup)
//   - Hyphenated objects: "click-tracks" -> "clickTrackId" (via schema lookup or camelCase construction)
//   - Exceptions: "multimedia" -> "imageId", "landing-pages-sites" -> "lpsId" (via schema lookup)
func parseRecordIdFromNode(c *Connector, objectName string, body *ajson.Node) string {
	// First, try to get the ID field name from the schemas (generated from swagger.json)
	// This handles ALL cases including exceptions like "multimedia" -> "imageId"
	idFieldName := getIDFieldNameFromSchemas(c, objectName)
	if recordID := tryExtractIDFromField(body, idFieldName); recordID != "" {
		return recordID
	}

	// Second: try constructing ID field name from objectName (e.g., "contacts" -> "contactId")
	// This handles most cases but may not work for hyphenated objects or exceptions
	constructedIDField := constructIDFieldFromObjectName(objectName)
	if recordID := tryExtractIDFromField(body, constructedIDField); recordID != "" {
		return recordID
	}

	// Last resort: try case-insensitive search for any field ending with "id"
	return extractIDFromMapPattern(body, objectName)
}

// tryExtractIDFromField attempts to extract an ID from a specific field name using jsonquery.TextWithDefault.
// Returns empty string if the field doesn't exist or extraction fails.
func tryExtractIDFromField(body *ajson.Node, fieldName string) string {
	if fieldName == "" {
		return ""
	}

	recordID, err := jsonquery.New(body).TextWithDefault(fieldName, "")
	if err != nil || recordID == "" {
		return ""
	}

	return recordID
}

// extractIDFromMapPattern extracts ID by pattern matching in the converted map.
// This is the last resort when schema lookup and construction both fail.
func extractIDFromMapPattern(body *ajson.Node, objectName string) string {
	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return ""
	}

	objectNameLower := strings.ToLower(objectName)
	singularLower := strings.ToLower(naming.NewSingularString(objectName).String())

	// First, try to find an ID field that matches the object name pattern
	if recordID := findMatchingIDField(data, objectNameLower, singularLower); recordID != "" {
		return recordID
	}

	// Absolute last resort: return the first field ending with "id"
	return findFirstIDField(data)
}

// findMatchingIDField searches for an ID field that matches the object name pattern.
func findMatchingIDField(data map[string]any, objectNameLower, singularLower string) string {
	for key, value := range data {
		keyLower := strings.ToLower(key)
		if strings.HasSuffix(keyLower, "id") {
			// Check if the key contains the object name (singular or plural) to avoid nested IDs
			if strings.Contains(keyLower, singularLower) || strings.Contains(keyLower, objectNameLower) {
				return convertIDToString(value)
			}
		}
	}

	return ""
}

// findFirstIDField returns the first field ending with "id" as a last resort.
func findFirstIDField(data map[string]any) string {
	for key, value := range data {
		if strings.HasSuffix(strings.ToLower(key), "id") {
			return convertIDToString(value)
		}
	}

	return ""
}

// constructIDFieldFromObjectName constructs an ID field name from the object name.
// This is a fallback when schema lookup fails. It handles:
// - Simple objects: "contacts" -> "contactId"
// - Hyphenated objects: "click-tracks" -> "clickTrackId" (camelCase conversion)
// Note: This may not work for exceptions like "multimedia" -> "imageId" or "landing-pages-sites" -> "lpsId".
func constructIDFieldFromObjectName(objectName string) string {
	// Convert to singular form
	singular := naming.NewSingularString(objectName).String()

	// Handle hyphenated objects: convert to camelCase
	// e.g., "click-track" -> "clickTrack"
	if strings.Contains(singular, "-") {
		parts := strings.Split(singular, "-")
		if len(parts) > 0 {
			var builder strings.Builder
			builder.WriteString(parts[0])
			for i := 1; i < len(parts); i++ {
				if len(parts[i]) > 0 {
					builder.WriteString(strings.ToUpper(parts[i][:1]) + parts[i][1:])
				}
			}
			builder.WriteString("Id")
			return builder.String()
		}
	}

	// For simple objects, just add "Id"

	return singular + "Id"
}

// getIDFieldNameFromSchemas returns the ID field name for a given object by looking it up
// in the schemas (which are generated from swagger.json).
// This ensures we use the same field names as defined in the API specification.
func getIDFieldNameFromSchemas(c *Connector, objectName string) string {
	// Try to get object metadata from schemas
	objectMetadata, err := metadata.Schemas.SelectOne(c.Module(), objectName)
	if err != nil {
		// If schema lookup fails, return empty string to fall back to other methods
		return ""
	}

	// Search through the fields to find one that ends with "Id" (case-insensitive)
	// GetResponse uses object-specific ID fields like contactId, campaignId, etc.
	for fieldName := range objectMetadata.Fields {
		fieldNameLower := strings.ToLower(fieldName)
		// Check if field name ends with "id" (e.g., contactId, campaignId)
		if strings.HasSuffix(fieldNameLower, "id") {
			// Prefer exact match over generic "id"
			if fieldNameLower != "id" {
				return fieldName
			}
		}
	}

	// If no specific ID field found, check for generic "id"
	if _, hasID := objectMetadata.Fields["id"]; hasID {
		return "id"
	}

	return ""
}

// convertIDToString converts various ID types to string.
func convertIDToString(idValue any) string {
	switch v := idValue.(type) {
	case string:
		return v
	case float64:
		return strconv.Itoa(int(v))
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return ""
	}
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	// Add RecordId to the path for delete operations
	// Format: DELETE /v3/{object}/{recordId}
	if params.RecordId != "" {
		url.AddPath(params.RecordId)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams, request *http.Request, response *common.JSONHTTPResponse) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{Success: true}, nil
}
