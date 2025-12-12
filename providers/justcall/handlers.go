package justcall

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// JustCall pagination: max 100 per page for most endpoints.
	// https://developer.justcall.io/reference/call_list_v21
	defaultPerPage = "100"
	perPage50      = "50"
	perPage20      = "20"

	// JustCall datetime format: yyyy-mm-dd hh:mm:ss or yyyy-mm-dd.
	// https://developer.justcall.io/reference/call_list_v21
	datetimeFormat = "2006-01-02 15:04:05"
)

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

	if !objectsWithoutPagination[params.ObjectName] {
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
	path, err := metadata.Schemas.LookupURLPath(c.ModuleID, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, path)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
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

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	// Determine the URL path
	modulePath := metadata.Schemas.LookupModuleURLPath(c.ModuleID)

	if specialPath, ok := objectsWithSpecialWritePath[params.ObjectName]; ok {
		url, err = urlbuilder.New(c.BaseURL, modulePath, specialPath)
	} else if objectsWithPathID[params.ObjectName] && params.RecordId != "" {
		// Objects like calls need ID in path: /calls/{id}
		path, pathErr := metadata.Schemas.FindURLPath(c.ModuleID, params.ObjectName)
		if pathErr != nil {
			return nil, pathErr
		}

		url, err = urlbuilder.New(c.BaseURL, modulePath, path, params.RecordId)
	} else {
		url, err = c.buildURL(params.ObjectName)
	}

	if err != nil {
		return nil, err
	}

	// Determine HTTP method
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
func extractRecordID(node *ajson.Node) string {
	// Try root level ID
	if id := tryExtractID(jsonquery.New(node)); id != "" {
		return id
	}

	// Try in data object
	if id := tryExtractID(jsonquery.New(node, "data")); id != "" {
		return id
	}

	// Try in nested data.data array (JustCall pattern for contacts)
	if dataArray, err := jsonquery.New(node, "data").ArrayOptional("data"); err == nil && len(dataArray) > 0 {
		return tryExtractID(jsonquery.New(dataArray[0]))
	}

	return ""
}

func tryExtractID(query *jsonquery.Query) string {
	if id, err := query.IntegerWithDefault("id", 0); err == nil && id != 0 {
		return strconv.FormatInt(id, 10)
	}

	if id, err := query.StrWithDefault("id", ""); err == nil && id != "" {
		return id
	}

	return ""
}
