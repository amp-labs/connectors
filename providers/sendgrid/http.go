package sendgrid

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// plainTextJSONClient wraps the authenticated HTTP client for SendGrid.
//
// Component readers/writers (operations.HTTPOperation) and common.JSONHTTPClient
// both parse successful responses via common.ParseJSONResponse, which only accepts
// application/json Content-Type. SendGrid sometimes returns a JSON body with
// text/plain (notably POST /v3/asm/groups). When the body is valid JSON, this
// client rewrites the Content-Type so parsing succeeds. Non-JSON text/plain is
// left unchanged; common.XMLHTTPClient covers XML, but there is no shared helper
// for other response formats.
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
