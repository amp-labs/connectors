package aircall

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "v1"

	// Pagination constants – see Aircall API pagination docs:
	// https://developer.aircall.io/api-references/#pagination
	//   - minimum: 1
	//   - default: 20
	//   - maximum: 50
	aircallMaxPerPage     = "50"
	aircallMaxPageSizeInt = 50
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Pagination
	pageSize := aircallMaxPerPage

	if params.PageSize > 0 {
		// Aircall maximum is 50
		if params.PageSize > aircallMaxPageSizeInt {
			pageSize = "50"
		} else {
			pageSize = strconv.Itoa(params.PageSize)
		}
	}

	url.WithQueryParam("per_page", pageSize)

	// Incremental sync: Add date range filters if provided
	// Aircall API uses Unix timestamps for 'from' and 'to' parameters
	// https://developer.aircall.io/api-references/#list-all-calls
	// Note: Not all objects support filtering by date (e.g. teams, tags).
	if supportsDateFiltering(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("from", strconv.FormatInt(params.Since.Unix(), 10))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("to", strconv.FormatInt(params.Until.Unix(), 10))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(params.ObjectName),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Aircall returns "meta" object with "next_page_link"
		return jsonquery.New(node, "meta").StrWithDefault("next_page_link", "")
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	// Marshal RecordData to JSON
	body, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	var url *urlbuilder.URL
	var method string

	if params.RecordId != "" {
		// Update
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, params.RecordId)
		if err != nil {
			return nil, err
		}

		switch params.ObjectName {
		case "contacts":
			// Aircall specific: POST for updates on contacts
			// https://developer.aircall.io/api-references/#update-a-contact
			method = http.MethodPost
		case "users", "tags", "numbers", "teams":
			// Standard PUT for updates
			method = http.MethodPut
		default:
			return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, "update not supported for "+params.ObjectName)
		}
	} else {
		// Create
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
		if err != nil {
			return nil, err
		}

		switch params.ObjectName {
		case "contacts", "users", "tags", "teams", "numbers", "calls":
			method = http.MethodPost
		default:
			return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, "create not supported for "+params.ObjectName)
		}
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	// Aircall responses are wrapped in a singular object key
	// e.g. "contacts" -> "contact"
	objectKey := naming.NewSingularString(params.ObjectName).String()

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	// Extract the object
	node, err := jsonquery.New(body).ObjectRequired(objectKey)
	if err != nil {
		return nil, err
	}

	// Extract ID - Aircall IDs are integers in JSON, but we need strings
	// Use TextWithDefault to convert automatically and fallback to the request ID if needed
	recordID, err := jsonquery.New(node).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, fmt.Errorf("failed to extract record ID: %w", err)
	}

	// Extract data
	dataMap, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     dataMap,
	}, nil
}

func supportsDateFiltering(objectName string) bool {
	switch objectName {
	case "calls", "users", "contacts", "numbers":
		return true
	case "teams", "tags":
		return false
	default:
		// Default to false to be safe
		return false
	}
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	// Validate that the object supports delete operations
	// Aircall supports delete for: contacts, users, tags, teams
	// See: https://developer.aircall.io/api-references/
	switch params.ObjectName {
	case "contacts", "users", "tags", "teams":
		// Supported
	default:
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, "delete not supported for "+params.ObjectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
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
