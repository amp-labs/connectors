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

// ErrorHandler allows the caller to inject their own HTTP error handling logic.
// All non-2xx responses will be passed to the error handler. If the error handler
// returns nil, then the error is ignored and the caller is responsible for handling
// the error. If the error handler returns an error, then that error is returned
// to the caller, as-is. Both the response and the response body are passed
// to the error handler as arguments.
type ErrorHandler func(rsp *http.Response, body []byte) error

// HTTPClient is an HTTP client which makes certain assumptions, such as
// that the response body is JSON. It also handles OAuth access token refreshes.
type HTTPClient struct {
	Base         string                  // optional base URL. If not set, then all URLs must be absolute.
	Client       AuthenticatedHTTPClient // underlying HTTP client. Required.
	ErrorHandler ErrorHandler            // optional error handler. If not set, then the default error handler is used.
}

func (c *HTTPClient) getURL(url string) (string, error) {
	return getURL(c.Base, url)
}
