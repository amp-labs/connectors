package calendly

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	DefaultPageSize = 100
	maxPageSize     = 100
)

// buildReadRequest constructs HTTP GET requests for reading Calendly data
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	// Ensure userURI is available for scheduled_events (lazy loading fallback)
	if params.ObjectName == "scheduled_events" && c.userURI == "" {
		_, err := c.GetPostAuthInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user info: %w", err)
		}
	}

	url, err := c.buildReadURL(ctx, params)
	if err != nil {
		return nil, err
	}

	return common.MakeJSONGetRequest(ctx, url.String(), nil)
}

// buildReadURL constructs the URL for reading Calendly data
func (c *Connector) buildReadURL(ctx context.Context, params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) > 0 {
		// Use the next page URL directly from pagination
		return urlbuilder.New(params.NextPage.String())
	}

	// Start building the base URL
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Set pagination parameters
	pageSize := DefaultPageSize
	url.WithQueryParam("count", strconv.Itoa(pageSize))

	// Add required user parameter for scheduled_events
	if params.ObjectName == "scheduled_events" {
		// Use the dynamically fetched user URI from GetPostAuthInfo
		url.WithQueryParam("user", c.userURI)
	}

	// Add time-based filtering if specified
	if !params.Since.IsZero() {
		url.WithQueryParam("min_start_time", params.Since.Format("2006-01-02T15:04:05Z"))
	}
	if !params.Until.IsZero() {
		url.WithQueryParam("max_start_time", params.Until.Format("2006-01-02T15:04:05Z"))
	}

	return url, nil
}

// parseReadResponse constructs HTTP GET requests for reading Calendly data
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// Use standard parsing for all objects
	result, err := common.ParseResult(
		response,
		common.ExtractRecordsFromPath("collection"),
		makeNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
	
	if err != nil {
		return nil, err
	}

	return result, nil
}

func makeNextRecordsURL(node *ajson.Node) (string, error) {
	pagination, err := jsonquery.New(node).ObjectOptional("pagination")
	if err != nil {
		return "", nil
	}

	if pagination == nil {
		return "", nil
	}

	nextPageURL, err := jsonquery.New(pagination).StringOptional("next_page")
	if err != nil {
		return "", nil
	}

	if nextPageURL == nil {
		return "", nil
	}

	return *nextPageURL, nil
}

// GetRecordsByIds retrieves specific records by their IDs
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	if len(recordIds) == 0 {
		return []common.ReadResultRow{}, nil
	}

	switch objectName {
	case "scheduled_events":
		return c.getScheduledEventsByIds(ctx, recordIds, fields)
	case "event_invitees":
		return c.getInviteesByIds(ctx, recordIds, fields)
	default:
		return nil, fmt.Errorf("GetRecordsByIds not supported for object type: %s", objectName)
	}
}

// getScheduledEventsByIds retrieves scheduled events by their IDs
func (c *Connector) getScheduledEventsByIds(
	ctx context.Context,
	recordIds []string,
	fields []string,
) ([]common.ReadResultRow, error) {
	var results []common.ReadResultRow

	for _, recordId := range recordIds {
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "scheduled_events", recordId)
		if err != nil {
			return nil, err
		}

		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			continue
		}

		body, ok := resp.Body()
		if !ok {
			continue
		}

		resource, err := jsonquery.New(body).ObjectRequired("resource")
		if err != nil {
			continue
		}

		fieldsMap, err := resource.GetObject()
		if err != nil {
			continue
		}
		
		fieldsData := make(map[string]any)
		for key, node := range fieldsMap {
			value, err := node.Value()
			if err == nil {
				fieldsData[key] = value
			}
		}

		row := common.ReadResultRow{
			Fields: fieldsData,
		}

		results = append(results, row)
	}

	return results, nil
}

// getInviteesByIds retrieves invitees by their IDs
func (c *Connector) getInviteesByIds(
	ctx context.Context,
	recordIds []string,
	fields []string,
) ([]common.ReadResultRow, error) {
	var results []common.ReadResultRow

	for _, recordId := range recordIds {
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "event_invitees", recordId)
		if err != nil {
			return nil, err
		}

		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			continue
		}

		body, ok := resp.Body()
		if !ok {
			continue
		}

		resource, err := jsonquery.New(body).ObjectRequired("resource")
		if err != nil {
			continue
		}

		fieldsMap, err := resource.GetObject()
		if err != nil {
			continue
		}
		
		fieldsData := make(map[string]any)
		for key, node := range fieldsMap {
			value, err := node.Value()
			if err == nil {
				fieldsData[key] = value
			}
		}

		row := common.ReadResultRow{
			Fields: fieldsData,
		}

		results = append(results, row)
	}

	return results, nil
} 