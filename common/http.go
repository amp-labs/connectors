package common

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string
	Value string
}

// InterpretError interprets the given HTTP response (in a fairly straightforward
// way) and returns an error that can be handled by the caller.
func InterpretError(res *http.Response, body []byte) error {
	switch res.StatusCode {
	case http.StatusUnauthorized:
		// Access token invalid, refresh token and retry
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrAccessToken, string(body)))
	case http.StatusForbidden:
		// Forbidden, not retryable
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrForbidden, string(body)))
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

func getURL(baseURL string, urlString string) (string, error) {
	if strings.HasPrefix(urlString, "http://") || strings.HasPrefix(urlString, "https://") {
		return urlString, nil
	}

	if len(baseURL) == 0 {
		return "", fmt.Errorf("%w (input is %q)", ErrEmptyBaseURL, urlString)
	}

	return url.JoinPath(baseURL, urlString)
}
