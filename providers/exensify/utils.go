package exensify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) executeRequest(ctx context.Context, body string) (*http.Response, error) {
	reqURL, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("requestJobDescription", body)
	encoded := form.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewBufferString(encoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient().Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Expensify content-Type is text even though it returns JSON,
	//  so we need to set it to application/json manually for the response to be parsed correctly
	resp.Header.Set("Content-Type", "application/json")

	return resp, nil
}

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

// checkResponseCode validates the responseCode field in an Expensify API response.
// Expensify returns HTTP 200 for all requests but signals failures via this field.
func checkResponseCode(result map[string]any) error {
	responseCode, ok := result["responseCode"]
	if !ok {
		return fmt.Errorf("response code is missing in the response: %w", common.ErrMissingExpectedValues)
	}

	//nolint:mnd
	if responseCode != float64(200) {
		responseMessage, ok := result["responseMessage"].(string)
		if !ok {
			responseMessage = "failed request with status code " + fmt.Sprint(responseCode)
		}

		return fmt.Errorf("%w: %s", common.ErrRequestFailed, responseMessage)
	}

	return nil
}

func buildReadBody(objectName string) (string, error) {
	body := map[string]any{
		"type": "get",
	}

	// We build the body based on the object
	if objectName == objectNamePolicy {
		body["inputSettings"] = map[string]any{
			"type":      "policyList",
			"adminOnly": false,
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(jsonBody), nil
}
