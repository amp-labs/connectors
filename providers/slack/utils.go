package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// jsonPostRequest builds an HTTP POST request with a JSON body.
func jsonPostRequest(ctx context.Context, rawURL string, body map[string]any) (*http.Request, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
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
