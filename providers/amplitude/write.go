package amplitude

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Write implements the WriteConnector interface for Amplitude.
// It handles both JSON and HTML/text responses from the API.
func (c *Connector) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	req, err := c.buildWriteRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient().Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, common.InterpretError(resp, bodyBytes)
	}

	contentType := resp.Header.Get("Content-Type")
	isJSON := strings.Contains(contentType, "application/json")

	// Try to parse as JSON if it's JSON content type or no content type specified
	if isJSON || len(contentType) == 0 {
		if result, err := c.tryParseAsJSON(params, bodyBytes); err == nil {
			return result, nil
		}
	}

	// If Content-Type is explicitly JSON but parsing failed, return an error
	if isJSON {
		return nil, common.ErrFailedToUnmarshalBody
	}

	// Handle non-JSON success (for Attribution object)
	return c.handleNonJSONSuccess(resp.StatusCode, bodyBytes), nil
}

func (c *Connector) tryParseAsJSON(params common.WriteParams, bodyBytes []byte) (*common.WriteResult, error) {
	if len(bodyBytes) == 0 {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	body, err := ajson.Unmarshal(bodyBytes)
	if err != nil {
		return nil, err
	}

	responseKey := writeObjectResponseField.Get(params.ObjectName)

	dataNode, err := jsonquery.New(body).ObjectOptional(responseKey)
	if err != nil {
		return nil, err
	}

	if dataNode == nil {
		// If object specific response key is not found, use the entire body
		dataNode = body
	}

	recordID, _ := jsonquery.New(dataNode).StrWithDefault("id", "")

	respMap, err := jsonquery.Convertor.ObjectToMap(dataNode)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     respMap,
	}, nil
}

// Attribution object returns response in html/text format.
func (c *Connector) handleNonJSONSuccess(statusCode int, bodyBytes []byte) *common.WriteResult {
	return &common.WriteResult{
		Success:  true,
		RecordId: "",
		Errors:   nil,
		Data: map[string]any{
			"status_code":   statusCode,
			"response_body": string(bodyBytes),
		},
	}
}
