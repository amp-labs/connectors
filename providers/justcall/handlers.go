package justcall

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
	"github.com/spyzhov/ajson"
)

// Pagination constants for JustCall API.
// Different endpoints have different max page sizes.
// Reference: https://developer.justcall.io/reference/pagination
const (
	apiVersion = "v2.1" // JustCall API version path

	defaultPerPage = "100" // Default max for most endpoints
	perPage50      = "50"  // For messages, whatsapp/messages, campaigns
	perPage20      = "20"  // For AI endpoints (calls_ai, meetings_ai)

	// JustCall datetime format for incremental sync.
	// Reference: https://developer.justcall.io/reference/call_list_v21
	datetimeFormat = "2006-01-02 15:04:05"
)

// ListObjectMetadata returns metadata for the requested objects, including custom fields.
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(common.ModuleRoot, objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		fields, err := c.requestCustomFields(ctx, objectName)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		// Attach fields to the object metadata.
		// Get a reference to the metadata in the map so changes are persisted.
		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			// Object not found in result, skip it
			continue
		}

		for _, field := range fields {
			objectMetadata.AddFieldMetadata(field.Label, common.FieldMetadata{
				DisplayName:  field.Label,
				ValueType:    field.getValueType(),
				ProviderType: field.Type,
				Values:       field.getValues(),
			})
		}

		// Write the modified metadata back to the map
		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}

// objectsWithoutPagination lists objects that don't support per_page parameter.
var objectsWithoutPagination = map[string]bool{ //nolint:gochecknoglobals
	"webhooks": true,
}

// objectsWithLowerPageLimit lists objects with lower per_page limits.
var objectsWithLowerPageLimit = map[string]string{ //nolint:gochecknoglobals
	"messages":               perPage50,
	"whatsapp/messages":      perPage50,
	"campaigns":              perPage50,
	"calls_ai":               perPage20,
	"meetings_ai":            perPage20,
	"sales_dialer/campaigns": perPage50,
}

// objectsWithIncrementalSync lists objects that support from_datetime/to_datetime filtering.
// https://developer.justcall.io/reference/call_list_v21
var objectsWithIncrementalSync = map[string]bool{ //nolint:gochecknoglobals
	"calls":                  true,
	"texts":                  true,
	"calls_ai":               true,
	"meetings_ai":            true,
	"sales_dialer/calls":     true,
	"whatsapp/messages":      true,
	"threads":                true,
	"sales_dialer/campaigns": true,
}

// buildReadRequest constructs the HTTP request for read operations.
// Handles pagination via next_page_link and incremental sync via from_datetime/to_datetime.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := c.buildURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if hasPagination := !objectsWithoutPagination[params.ObjectName]; hasPagination {
		perPage := defaultPerPage
		if limit, ok := objectsWithLowerPageLimit[params.ObjectName]; ok {
			perPage = limit
		}

		url.WithQueryParam("per_page", perPage)
	}

	// Add incremental sync parameters if supported and provided.
	if objectsWithIncrementalSync[params.ObjectName] {
		if !params.Since.IsZero() {
			url.WithQueryParam("from_datetime", params.Since.Format(datetimeFormat))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("to_datetime", params.Until.Format(datetimeFormat))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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
	return common.ParseResult(
		response,
		makeGetRecords(),
		makeNextRecordsURL(),
		common.MakeMarshaledDataFunc(flattenCustomFields),
		params.Fields,
	)
}

func makeGetRecords() common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional("data")
	}
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node).StrWithDefault("next_page_link", "")
	}
}

// objectsWithPathID lists objects where RecordId goes in the URL path for updates.
var objectsWithPathID = map[string]bool{ //nolint:gochecknoglobals
	"calls": true,
}

// objectsWithSpecialWritePath maps objects to special write endpoints (not in metadata).
var objectsWithSpecialWritePath = map[string]string{ //nolint:gochecknoglobals
	"texts":                          "/texts/new",
	"contacts/status":                "/contacts/status",
	"texts/threads/tag":              "/texts/threads/tag",
	"sales_dialer/campaigns/contact": "/sales_dialer/campaigns/contact",
	"voice-agents/calls":             "/voice-agents/calls",
	"users/availability":             "/users/availability",
}

// objectsWithPUTOnly lists objects that always use PUT (even without RecordId).
var objectsWithPUTOnly = map[string]bool{ //nolint:gochecknoglobals
	"contacts/status":    true,
	"users/availability": true,
}

// buildWriteRequest constructs the HTTP request for write operations.
// Uses POST for create, PUT for update. Some objects use special endpoints.
// Reference: https://developer.justcall.io/reference/introduction
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if params.RecordId != "" || objectsWithPUTOnly[params.ObjectName] {
		method = http.MethodPut
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

	return req, nil
}

// buildWriteURL constructs the URL for write operations.
// Handles special paths, path-based IDs, and standard metadata paths.
func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, error) {
	// Check for special write paths (e.g., /texts/new, /contacts/status)
	if specialPath, ok := objectsWithSpecialWritePath[params.ObjectName]; ok {
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, specialPath)
	}

	// Objects like calls need ID in path: /calls/{id}
	if objectsWithPathID[params.ObjectName] && params.RecordId != "" {
		path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, params.ObjectName)
		if err != nil {
			return nil, err
		}

		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path, params.RecordId)
	}

	return c.buildURL(params.ObjectName)
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	// Try to extract record ID from response
	recordID := params.RecordId
	if recordID == "" {
		recordID = extractRecordID(node)
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil { //nolint:nilerr
		return &common.WriteResult{
			Success:  true,
			RecordId: recordID,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

// extractRecordID extracts the record ID from response, handling both string and numeric IDs.
// JustCall has different response structures:
//   - Tags: {"status": "success", "data": {"id": 123, ...}}
//   - Contacts: {"status": "success", "data": [{"id": 123, ...}]}.
func extractRecordID(node *ajson.Node) string {
	query := jsonquery.New(node)

	// Try root level ID
	if id, err := query.TextWithDefault("id", ""); err == nil && id != "" {
		return id
	}

	// Try in data object (for tags: data.id)
	if dataNode, err := query.ObjectOptional("data"); err == nil && dataNode != nil {
		if id, err := jsonquery.New(dataNode).TextWithDefault("id", ""); err == nil && id != "" {
			return id
		}
	}

	// Try in data array (for contacts: data[0].id)
	if dataArray, err := query.ArrayOptional("data"); err == nil && len(dataArray) > 0 {
		if id, err := jsonquery.New(dataArray[0]).TextWithDefault("id", ""); err == nil && id != "" {
			return id
		}
	}

	return ""
}

// deletableObjectsWithPathID lists objects where RecordId goes in the URL path for delete.
var deletableObjectsWithPathID = map[string]string{ //nolint:gochecknoglobals
	"tags":                  "/texts/tags",
	"sales_dialer/contacts": "/sales_dialer/contacts",
	"webhooks":              "/webhooks/url",
}

// deletableObjectsWithQueryID lists objects where RecordId goes in query params for delete.
// Note: JustCall API uses query parameters for contacts delete (e.g., /contacts?id=12345)
// instead of the more common path-based pattern (/contacts/12345).
var deletableObjectsWithQueryID = map[string]string{ //nolint:gochecknoglobals
	"contacts": "id",
}

// buildDeleteRequest constructs the HTTP request for delete operations.
// Note: JustCall requires empty JSON body {} for DELETE requests.
// Two patterns: path-based ID (/endpoint/{id}) or query param (/endpoint?id={id}).
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.buildDeleteURL(params)
	if err != nil {
		return nil, err
	}

	// JustCall requires a JSON body even for DELETE requests
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) buildDeleteURL(params common.DeleteParams) (*urlbuilder.URL, error) {
	// Check if object uses path-based ID
	if basePath, ok := deletableObjectsWithPathID[params.ObjectName]; ok {
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, basePath, params.RecordId)
	}

	// Check if object uses query param for ID
	queryParam, ok := deletableObjectsWithQueryID[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(queryParam, params.RecordId)

	return url, nil
}

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
