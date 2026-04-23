package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/okta/metadata"
	"github.com/spyzhov/ajson"
)

const (
	limitKey  = "limit"
	pageLimit = 200 // Okta maximum per page
	filterKey = "filter"
	sinceKey  = "since"
)

// Objects supporting incremental sync via provider-side lastUpdated filter.
// Only users support the filter=lastUpdated gt "..." query parameter.
// Reference: https://developer.okta.com/docs/reference/api/users/#list-users-with-a-filter
//
// nolint:gochecknoglobals
var objectsWithProviderSideFilter = datautils.NewStringSet(
	"users",
)

// Objects that support connector-side filtering via lastUpdated field.
// These objects have a lastUpdated timestamp but don't support provider-side filtering.
// Groups and apps return 400 "Invalid search criteria" when using lastUpdated filter.
//
// nolint:gochecknoglobals
var objectsWithConnectorSideFilter = datautils.NewStringSet(
	"groups",
	"apps",
	"devices",
	"idps",
	"authorizationServers",
	"trustedOrigins",
	"zones",
	"authenticators",
	"policies",
	"eventHooks",
)

// responseField returns the JSON path for extracting records.
// Most Okta endpoints return arrays at root level, except domains.
func responseField(objectName string) string {
	// Domains endpoint wraps the array in a "domains" key
	if objectName == "domains" {
		return "domains"
	}

	// Empty string means the response is an array at root level
	return ""
}

// buildReadRequest constructs the HTTP request for read operations.
// Reference: https://developer.okta.com/docs/api/
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	// Use NextPage directly (from Link header) if provided
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	// Build URL from metadata
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(limitKey, readhelper.PageSizeWithDefaultStr(params, strconv.Itoa(pageLimit)))

	// Add incremental sync filter based on object type
	if !params.Since.IsZero() {
		if params.ObjectName == "logs" {
			// Logs API uses 'since' query param instead of filter expression
			// Reference: https://developer.okta.com/docs/reference/api/system-log/#request-parameters
			url.WithQueryParam(sinceKey, datautils.Time.FormatRFC3339inUTC(params.Since))
		} else if objectsWithProviderSideFilter.Has(params.ObjectName) {
			// Users support the lastUpdated filter expression for incremental sync.
			// Okta requires %20 for spaces in filter expressions, not +.
			// Reference: https://developer.okta.com/docs/reference/api/users/#list-users-with-a-filter
			filterValue := "lastUpdated gt \"" + datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since) + "\""
			url.WithQueryParam(filterKey, filterValue)
			url.AddEncodingExceptions(map[string]string{"+": "%20"})
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// parseReadResponse parses the HTTP response from read operations.
// Okta uses Link headers for pagination (cursor-based).
// Custom profile fields are flattened to root level for users and groups.
// Reference: https://developer.okta.com/docs/api/#pagination
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// Use flattenProfileFields for objects with profile data (users, groups)
	// This moves custom fields from profile.{field} to root level
	var transformer common.RecordTransformer
	if objectsWithCustomFields.Has(params.ObjectName) {
		transformer = flattenProfileFields
	}

	return common.ParseResultFiltered(
		params,
		response,
		common.MakeRecordsFunc(responseField(params.ObjectName)),
		makeFilterFunc(params, response.Headers),
		common.MakeMarshaledDataFunc(transformer),
		params.Fields,
	)
}

// makeFilterFunc returns the appropriate filter function based on object type.
// Users and logs use provider-side filtering and don't need connector-side filtering.
// Objects with lastUpdated field but no provider-side support use connector-side filtering.
func makeFilterFunc(params common.ReadParams, headers http.Header) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL(headers)

	// Objects with provider-side filtering don't need connector-side filtering
	if objectsWithProviderSideFilter.Has(params.ObjectName) || params.ObjectName == "logs" {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	// Objects without any timestamp field - no filtering possible
	if !objectsWithConnectorSideFilter.Has(params.ObjectName) {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	// Apply connector-side filtering using lastUpdated field
	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		"lastUpdated",
		time.RFC3339,
		nextPageFunc,
	)
}

// makeNextRecordsURL extracts the next page URL from Link header.
// Okta uses Link headers with rel="next" for pagination.
// Reference: https://developer.okta.com/docs/api/#link-header
func makeNextRecordsURL(responseHeaders http.Header) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(&common.JSONHTTPResponse{Headers: responseHeaders}, "next"), nil
	}
}

// buildWriteRequest constructs the HTTP request for write operations.
// POST is used for creates, PUT for updates (except users which use POST for partial updates).
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		// Users use POST for partial updates, other objects use PUT for full replacement.
		if params.ObjectName != "users" {
			method = http.MethodPut
		}
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// parseWriteResponse parses the HTTP response from write operations.
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

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

// buildDeleteRequest constructs the HTTP request for delete operations.
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

// parseDeleteResponse parses the HTTP response from delete operations.
func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	return &common.DeleteResult{
		Success: true,
	}, nil
}
