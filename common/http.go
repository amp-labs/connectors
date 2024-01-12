package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string
	Value string
}

// ErrorHandler allows the caller to inject their own HTTP error handling logic.
// All non-2xx responses will be passed to the error handler. If the error handler
// returns nil, then the error is ignored and the caller is responsible for handling
// the error. If the error handler returns an error, then that error is returned
// to the caller, as-is. Both the response and the response body are passed
// to the error handler as arguments.
type ErrorHandler func(rsp *http.Response, body []byte) error

// HTTPClient is an HTTP client that handles OAuth access token refreshes.
type HTTPClient struct {
	Base         string                  // optional base URL. If not set, then all URLs must be absolute.
	Client       AuthenticatedHTTPClient // underlying HTTP client. Required.
	ErrorHandler ErrorHandler            // optional error handler. If not set, then the default error handler is used.
}

// getURL returns the base prefixed URL.
func (h *HTTPClient) getURL(url string) (string, error) {
	return getURL(h.Base, url)
}

// Get makes a GET request to the given URL and returns the response. If the response is not a 2xx,
// an error is returned. If the response is a 401, the caller should refresh the access token
// and retry the request. If errorHandler is nil, then the default error handler is used.
// If not, the caller can inject their own error handling logic.
func (h *HTTPClient) Get(ctx context.Context, url string, headers []Header) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}

	// Make the request, get the response body
	res, body, err := h.httpGet(ctx, fullURL, headers) //nolint:bodyclose
	if err != nil {
		return nil, nil, err
	}

	return res, body, nil
}

// Post makes a POST request to the given URL and returns the response & response body.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request. If errorHandler is nil, then the default error
// handler is used. If not, the caller can inject their own error handling logic.
func (h *HTTPClient) Post(ctx context.Context,
	url string, reqBody any, headers []Header,
) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}

	// Make the request, get the response body
	res, body, err := h.httpPost(ctx, fullURL, headers, reqBody) //nolint:bodyclose
	if err != nil {
		return nil, nil, err
	}

	return res, body, nil
}

// Patch makes a PATCH request to the given URL and returns the response & response body.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request. If errorHandler is nil, then the default error
// handler is used. If not, the caller can inject their own error handling logic.
func (h *HTTPClient) Patch(ctx context.Context,
	url string, reqBody any, headers []Header,
) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}
	// Make the request, get the response body
	res, body, err := h.httpPatch(ctx, fullURL, headers, reqBody) //nolint:bodyclose
	if err != nil {
		return nil, nil, err
	}

	return res, body, nil
}

func (h *HTTPClient) Put(ctx context.Context,
	url string, reqBody any, headers []Header,
) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}
	// Make the request, get the response body
	res, body, err := h.httpPut(ctx, fullURL, headers, reqBody) //nolint:bodyclose
	if err != nil {
		return nil, nil, err
	}

	return res, body, nil
}

func (h *HTTPClient) Delete(ctx context.Context,
	url string, headers []Header,
) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}
	// Make the request, get the response body
	res, body, err := h.httpDelete(ctx, fullURL, headers) //nolint:bodyclose
	if err != nil {
		return nil, nil, err
	}

	return res, body, nil
}

// httpGet makes a GET request to the given URL and returns the response & response body.
func (h *HTTPClient) httpGet(ctx context.Context,
	url string, headers []Header,
) (*http.Response, []byte, error) {
	req, err := makeGetRequest(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}

	return h.sendRequest(req)
}

// httpPost makes a POST request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPost(ctx context.Context, url string,
	headers []Header, body any,
) (*http.Response, []byte, error) {
	req, err := makePostRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	return h.sendRequest(req)
}

// httpPatch makes a PATCH request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPatch(ctx context.Context,
	url string, headers []Header, body any,
) (*http.Response, []byte, error) {
	req, err := makePatchRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	return h.sendRequest(req)
}

// httpPut makes a PUT request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPut(ctx context.Context,
	url string, headers []Header, body any,
) (*http.Response, []byte, error) {
	req, err := makePutRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	return h.sendRequest(req)
}

// httpDelete makes a DELETE request to the given URL and returns the response & response body.
func (h *HTTPClient) httpDelete(ctx context.Context,
	url string, headers []Header,
) (*http.Response, []byte, error) {
	req, err := makeDeleteRequest(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}

	return h.sendRequest(req)
}

// makeGetRequest creates a GET request with the given headers.
func makeGetRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return addHeaders(req, headers)
}

// makePostRequest creates a POST request with the given headers and body, and adds the
// Content-Type header. It then returns the request.
func makePostRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "application/json"})
	req.ContentLength = int64(len(jBody))

	return addHeaders(req, headers)
}

// makePatchRequest creates a PATCH request with the given headers and body, and adds the
// Content-Type header. It then returns the request.
func makePatchRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "application/json"})
	req.ContentLength = int64(len(jBody))

	return addHeaders(req, headers)
}

// makePutRequest creates a PUT request with the given headers and body, and adds the
// Content-Type header. It then returns the request.
func makePutRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "application/json"})
	req.ContentLength = int64(len(jBody))

	return addHeaders(req, headers)
}

// makeDeleteRequest creates a DELETE request with the given headers. It then returns the request.
func makeDeleteRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return addHeaders(req, headers)
}

// sendRequest sends the given request and returns the response & response body.
func (h *HTTPClient) sendRequest(req *http.Request) (*http.Response, []byte, error) {
	// Send the request
	res, err := h.Client.Do(req)
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
		if h.ErrorHandler != nil {
			return nil, nil, h.ErrorHandler(res, body)
		}

		return nil, nil, InterpretError(res, body)
	}

	return res, body, nil
}

// getURL returns the given URL if it is an absolute URL, or the given URL joined with the base URL.
func getURL(baseURL string, urlString string) (string, error) {
	if strings.HasPrefix(urlString, "http://") || strings.HasPrefix(urlString, "https://") {
		return urlString, nil
	}

	if len(baseURL) == 0 {
		return "", fmt.Errorf("%w (input is %q)", ErrEmptyBaseURL, urlString)
	}

	return url.JoinPath(baseURL, urlString)
}

// addHeaders adds the given headers to the request.
func addHeaders(req *http.Request, headers []Header) (*http.Request, error) {
	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req, nil
}
