package sendgrid

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// plainTextJSONClient wraps the authenticated client so SendGrid responses that
// return JSON with Content-Type text/plain (notably POST /v3/asm/groups) are
// accepted by common.ParseJSONResponse.
type plainTextJSONClient struct {
	common.AuthenticatedHTTPClient
}

func (c *plainTextJSONClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.AuthenticatedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	return normalizeJSONContentType(resp)
}

func normalizeJSONContentType(resp *http.Response) (*http.Response, error) {
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		return resp, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	if !json.Valid(body) {
		return resp, nil
	}

	resp.Header.Set("Content-Type", "application/json")

	return resp, nil
}
