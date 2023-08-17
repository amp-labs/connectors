package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/spyzhov/ajson"
)

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string
	Value string
}

// ErrNotJSON is returned when a response is not JSON.
var ErrNotJSON = errors.New("response is not JSON")

// GetJSON makes a GET request to the given URL and returns the response body as a JSON object.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request.
func GetJSON(ctx context.Context, client *http.Client, url string, headers ...Header) (*ajson.Node, error) {
	// Make the request, get the response body
	res, body, err := httpGet(ctx, client, url, headers) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	// Ensure the response is JSON
	ct := res.Header.Get("Content-Type")
	if len(ct) > 0 {
		mimeType, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content type: %w", err)
		}

		if mimeType != "application/json" {
			return nil, fmt.Errorf("%w: expected content type to be application/json, got %s", ErrNotJSON, mimeType)
		}
	}

	// Unmarshall the response body into JSON
	jsonBody, err := ajson.Unmarshal(body)
	if err != nil {
		return nil, NewHTTPStatusError(res.StatusCode, fmt.Errorf("failed to unmarshall response body into JSON: %w", err))
	}

	return jsonBody, nil
}

func httpGet(ctx context.Context, client *http.Client, url string, headers []Header) (*http.Response, []byte, error) {
	// Create a new GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	// Request JSON
	req.Header.Add("Accept", "application/json")

	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	// Send the request
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error sending request: %w", err)
	}

	// Read the response body
	body, err := io.ReadAll(res.Body)

	defer func() {
		if res != nil && res.Body != nil {
			if closeErr := res.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	if err != nil {
		return nil, nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check the response status code
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, nil, interpretError(res, body)
	}

	return res, body, nil
}

func interpretError(res *http.Response, body []byte) error {
	switch res.StatusCode {
	case http.StatusUnauthorized:
		// Access token invalid, refresh token and retry
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrAccessToken, string(body)))
	case http.StatusForbidden:
		// Forbidden, treat this as the API being disabled
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrApiDisabled, string(body)))
	case http.StatusNotFound:
		// Semantics are debatable (temporarily missing vs. permanently gone), but for now treat this as a retryable error
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: entity not found (%s)", ErrRetryable, string(body)))
	case http.StatusTooManyRequests:
		// Too many requests, sleep and then retry
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrRetryable, string(body)))
	}

	if res.StatusCode >= 400 && res.StatusCode < 500 {
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrCaller, string(body)))
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrServer, string(body)))
	}

	return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrUnknown, string(body)))
}
