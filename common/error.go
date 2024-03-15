package common

import (
	"fmt"
	"net/http"
)

var ErrNoProxyConnectorFound = fmt.Errorf("no proxy connector found")

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
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrRetryable, string(body)))
	case http.StatusTooManyRequests:
		// Too many requests, retryable
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrRetryable, string(body)))
	}

	if res.StatusCode >= 400 && res.StatusCode < 500 {
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrCaller, string(body)))
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrServer, string(body)))
	}

	return NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", ErrUnknown, string(body)))
}
