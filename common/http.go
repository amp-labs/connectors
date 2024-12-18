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

	"github.com/amp-labs/connectors/common/logging"
)

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ErrorHandler allows the caller to inject their own HTTP error handling logic.
// All non-2xx responses will be passed to the error handler. If the error handler
// returns nil, then the error is ignored and the caller is responsible for handling
// the error. If the error handler returns an error, then that error is returned
// to the caller, as-is. Both the response and the response body are passed
// to the error handler as arguments.
type ErrorHandler func(rsp *http.Response, body []byte) error

type ResponseHandler func(rsp *http.Response) (*http.Response, error)

// HTTPClient is an HTTP client that handles OAuth access token refreshes.
type HTTPClient struct {
	Base            string                  // optional base URL. If not set, then all URLs must be absolute.
	Client          AuthenticatedHTTPClient // underlying HTTP client. Required.
	ErrorHandler    ErrorHandler            // optional error handler. If not set, then the default error handler is used.
	ResponseHandler ResponseHandler         // optional, Allows mutation of the http.Response from the Saas API response.
}

// getURL returns the base prefixed URL.
func (h *HTTPClient) getURL(url string) (string, error) {
	return getURL(h.Base, url)
}

// redactSensitiveHeaders redacts sensitive headers from the given headers.
func redactSensitiveHeaders(hdrs []Header) []Header {
	if hdrs == nil {
		return nil
	}

	redacted := make([]Header, 0, len(hdrs))

	for _, hdr := range hdrs {
		switch {
		case strings.EqualFold(hdr.Key, "Authorization"):
			redacted = append(redacted, Header{Key: hdr.Key, Value: "<redacted>"})
		case strings.EqualFold(hdr.Key, "Proxy-Authorization"):
			redacted = append(redacted, Header{Key: hdr.Key, Value: "<redacted>"})
		case strings.EqualFold(hdr.Key, "x-amz-security-token"):
			redacted = append(redacted, Header{Key: hdr.Key, Value: "<redacted>"})
		case strings.EqualFold(hdr.Key, "X-Api-Key"):
			redacted = append(redacted, Header{Key: hdr.Key, Value: "<redacted>"})
		case strings.EqualFold(hdr.Key, "X-Admin-Key"):
			redacted = append(redacted, Header{Key: hdr.Key, Value: "<redacted>"})
		default:
			redacted = append(redacted, hdr)
		}
	}

	return redacted
}

// Get makes a GET request to the given URL and returns the response. If the response is not a 2xx,
// an error is returned. If the response is a 401, the caller should refresh the access token
// and retry the request. If errorHandler is nil, then the default error handler is used.
// If not, the caller can inject their own error handling logic.
func (h *HTTPClient) Get(ctx context.Context, url string, headers ...Header) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}

	logging.Logger(ctx).Debug("HTTP request",
		"method", "GET", "url", fullURL,
		"headers", redactSensitiveHeaders(headers))

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
	url string, reqBody []byte, headers ...Header,
) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}

	logging.Logger(ctx).Debug("HTTP request",
		"method", "POST", "url", fullURL,
		"headers", redactSensitiveHeaders(headers),
		"bodySize", len(reqBody))

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
	url string, reqBody any, headers ...Header,
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
	url string, reqBody any, headers ...Header,
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
	url string, headers ...Header,
) (*http.Response, []byte, error) {
	fullURL, err := h.getURL(url)
	if err != nil {
		return nil, nil, err
	}

	logging.Logger(ctx).Debug("HTTP request",
		"method", "DELETE", "url", fullURL,
		"headers", redactSensitiveHeaders(headers))

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
	req, err := MakeGetRequest(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}

	return h.sendRequest(req)
}

// httpPost makes a POST request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPost(ctx context.Context, url string,
	headers []Header, body []byte,
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

// MakeGetRequest creates a GET request with the given headers.
func MakeGetRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return addHeaders(req, headers), nil
}

// makePostRequest creates request that will post bytes of data. If no content type defaults to JSON.
func makePostRequest(ctx context.Context, url string, headers []Header, data []byte) (*http.Request, error) {
	buffer := bytes.NewBuffer(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buffer)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.ContentLength = int64(len(data))

	return addJSONContentTypeIfNotPresent(addHeaders(req, headers)), nil
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

	req.ContentLength = int64(len(jBody))

	logging.Logger(ctx).Debug("HTTP request",
		"method", "PATCH", "url", url,
		"headers", redactSensitiveHeaders(headers),
		"bodySize", len(jBody))

	return addJSONContentTypeIfNotPresent(addHeaders(req, headers)), nil
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

	req.ContentLength = int64(len(jBody))

	logging.Logger(ctx).Debug("HTTP request",
		"method", "PUT", "url", url,
		"headers", redactSensitiveHeaders(headers),
		"bodySize", len(jBody))

	return addJSONContentTypeIfNotPresent(addHeaders(req, headers)), nil
}

// makeDeleteRequest creates a DELETE request with the given headers. It then returns the request.
func makeDeleteRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return addHeaders(req, headers), nil
}

// sendRequest sends the given request and returns the response & response body.
func (h *HTTPClient) sendRequest(req *http.Request) (*http.Response, []byte, error) { //nolint:cyclop
	// Send the request
	res, err := h.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// Apply the ResponseHandler if provided
	if h.ResponseHandler != nil {
		res, err = h.ResponseHandler(res)
		if err != nil {
			return nil, nil, err
		}
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
func addHeaders(req *http.Request, headers []Header) *http.Request {
	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req
}

// addJSONContentTypeIfNotPresent adds the Content-Type header if it is not already present.
func addJSONContentTypeIfNotPresent(req *http.Request) *http.Request {
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/json")
	}

	return req
}

func GetResponseBodyOnce(response *http.Response) []byte {
	defer func() {
		if response != nil && response.Body != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Error reading response body", "error", err)

		return nil
	}

	return body
}
