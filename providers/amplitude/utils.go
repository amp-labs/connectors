package amplitude

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

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

func (c *Connector) constructURL(objectName string) (*urlbuilder.URL, error) {
	apiVersion := objectAPIVersion.Get(objectName)

	path := fmt.Sprintf("api/%s/%s", apiVersion, objectName)

	if objectName == objectNameEvents {
		path = fmt.Sprintf("api/%s/%s/list", apiVersion, objectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	return url, nil
}

// Helper to extract API key (username) from the auth client.
func extractAPIKey(client common.AuthenticatedHTTPClient, ctx context.Context) (string, error) {
	// Use a safe dummy endpoint (e.g., Amplitude's health check or base URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.amplitude.com/health", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	auth := resp.Request.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Basic ") {
		return "", errors.New("no basic auth found") //nolint: err113
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(string(decoded), ":", 2) //nolint:mnd
	if len(parts) != 2 {                             //nolint:mnd
		return "", errors.New("invalid auth format") //nolint: err113
	}

	return parts[0], nil
}
