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

	resp.Header.Set("Content-Type", "application/json")

	// var result map[string]any

	// if err = json.Unmarshal(common.GetResponseBodyOnce(resp), &result); err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	// }

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

func buildReadBody(objectName string) (string, error) {

	body := map[string]any{
		"type": "get",
	}

	if objectName == objectNamePolicy {
		body["inputSettings"] = map[string]any{
			"type":      "policyList",
			"adminOnly": true,
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(jsonBody), nil
}
