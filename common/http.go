// nolint:revive
package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common/logging"
	"github.com/google/uuid"
)

// HeaderMode determines how the header should be applied to the request.
type HeaderMode int

const (
	// headerModeUnset is the default mode. It appends the header to the request.
	headerModeUnset = iota

	// HeaderModeAppend appends the header to the request.
	HeaderModeAppend

	// HeaderModeOverwrite unconditionally overwrites the header in the request.
	HeaderModeOverwrite

	// HeaderModeSetIfMissing sets the header in the request if it is not already set.
	HeaderModeSetIfMissing
)

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string     `json:"key"`
	Value string     `json:"value"`
	Mode  HeaderMode `json:"mode"`
}

func (h Header) ApplyToRequest(req *http.Request) {
	switch h.Mode {
	case HeaderModeOverwrite:
		req.Header.Set(h.Key, h.Value)
	case HeaderModeSetIfMissing:
		if len(req.Header.Values(h.Key)) == 0 {
			req.Header.Add(h.Key, h.Value)
		}
	case HeaderModeAppend:
		fallthrough
	case headerModeUnset:
		fallthrough
	default:
		req.Header.Add(h.Key, h.Value)
	}
}

func (h Header) String() string {
	return fmt.Sprintf("%s: %s", h.Key, h.Value)
}

func (h Header) equals(other Header) bool {
	return textproto.CanonicalMIMEHeaderKey(h.Key) == textproto.CanonicalMIMEHeaderKey(other.Key) &&
		h.Value == other.Value &&
		h.Mode == other.Mode
}

var HeaderFormURLEncoded = Header{ // nolint:gochecknoglobals
	Key:   "Content-Type",
	Value: "application/x-www-form-urlencoded",
}

type Headers []Header

func (h Headers) Has(target Header) bool {
	for _, header := range h {
		if header.equals(target) {
			return true
		}
	}

	return false
}

func (h Headers) ApplyToRequest(req *http.Request) {
	for _, header := range h {
		header.ApplyToRequest(req)
	}
}

func (h Headers) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(h))

	for _, header := range h {
		attrs = append(attrs, slog.String(header.Key, header.Value))
	}

	return slog.GroupValue(attrs...)
}

// ErrorHandler allows the caller to inject their own HTTP error handling logic.
// All non-2xx responses will be passed to the error handler. If the error handler
// returns nil, then the error is ignored and the caller is responsible for handling
// the error. If the error handler returns an error, then that error is returned
// to the caller, as-is. Both the response and the response body are passed
// to the error handler as arguments.
type ErrorHandler func(rsp *http.Response, body []byte) error

type ResponseHandler func(rsp *http.Response) (*http.Response, error)

// ShouldHandleError determines whether the default or custom ErrorHandler
// should be invoked for a given HTTP response.
// Returning true indicates that the response represents an error that requires handling.
type ShouldHandleError func(response *http.Response) bool

// HTTPClient is an HTTP client that handles OAuth access token refreshes
// and provides hooks for custom error and response handling.
type HTTPClient struct {
	// [Deprecated] URL endpoints are not the responsibility of HTTPClient.
	// NOTE: to avoid linter errors the deprecation comment is not of correct golang formatting.
	// Optional base URL. If unset, all request URLs must be absolute.
	Base string
	// Underlying HTTP client. Required.
	Client AuthenticatedHTTPClient
	// Optional ErrorHandler. If not set, then the default error handler is used.
	ErrorHandler ErrorHandler
	// Optional ResponseHandler, allowing mutation of the http.Response returned by the SaaS API.
	ResponseHandler ResponseHandler
	// Optional predicate deciding whether the ErrorHandler should be invoked.
	ShouldHandleError ShouldHandleError
}

// getURL returns the base prefixed URL.
func (h *HTTPClient) getURL(url string) (string, error) { // nolint:funcorder
	return getURL(h.Base, url)
}

// redactSensitiveRequestHeaders redacts sensitive headers from the given headers.
func redactSensitiveRequestHeaders(hdrs []Header) Headers {
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

func redactSensitiveResponseHeaders(hdrs []Header) Headers {
	if hdrs == nil {
		return nil
	}

	redacted := make([]Header, 0, len(hdrs))

	for _, hdr := range hdrs {
		switch {
		case strings.EqualFold(hdr.Key, "Set-Cookie"):
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

	// Make the request, get the response body
	res, body, err := h.httpDelete(ctx, fullURL, headers) //nolint:bodyclose
	if err != nil {
		return nil, nil, err
	}

	return res, body, nil
}

// httpGet makes a GET request to the given URL and returns the response & response body.
func (h *HTTPClient) httpGet(ctx context.Context, //nolint:dupl
	url string, headers []Header,
) (*http.Response, []byte, error) {
	req, err := MakeGetRequest(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}

	correlationId := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		logRequestWithoutBody(logging.VerboseLogger(ctx), req, "GET", correlationId, url)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, "GET", correlationId, url)
	}

	rsp, body, err := h.sendRequest(req)

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, "GET", correlationId, url, body)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, "GET", correlationId, url)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", "GET", "url", url,
			"correlationId", correlationId, "error", err)

		return nil, nil, err
	}

	return rsp, body, nil
}

// httpPost makes a POST request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPost(ctx context.Context, url string, //nolint:dupl
	headers []Header, requestBody []byte,
) (*http.Response, []byte, error) {
	req, err := makePostRequest(ctx, url, headers, requestBody)
	if err != nil {
		return nil, nil, err
	}

	correlationId := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		// body is nil because makePostRequest sometimes returns an altered body reader
		// so we rely on req and not body for logging purposes. If we use body, it will
		// not just log the wrong thing, it will swap the body out, which will lead to
		// unexpected behavior in the request.
		logRequestWithBody(logging.VerboseLogger(ctx), req, "POST", correlationId, url, nil)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, "POST", correlationId, url)
	}

	rsp, responseBody, err := h.sendRequest(req)

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, "POST", correlationId, url, responseBody)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, "POST", correlationId, url)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", "POST", "url", url,
			"correlationId", correlationId, "error", err)

		return nil, nil, err
	}

	return rsp, responseBody, nil
}

// httpPatch makes a PATCH request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPatch(ctx context.Context, //nolint:dupl
	url string, headers []Header, body any,
) (*http.Response, []byte, error) {
	req, err := makePatchRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	correlationId := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		var serializedBody []byte

		if body != nil {
			serializedBody, err = json.Marshal(body)
			if err != nil {
				logging.Logger(ctx).Error("Failed to serialize request body",
					"method", "PATCH", "url", url,
					"correlationId", correlationId, "error", err)
			}
		}

		logRequestWithBody(logging.VerboseLogger(ctx), req, "PATCH", correlationId, url, serializedBody)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, "PATCH", correlationId, url)
	}

	rsp, rspBody, err := h.sendRequest(req)

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, "PATCH", correlationId, url, rspBody)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, "PATCH", correlationId, url)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", "PATCH", "url", url,
			"correlationId", correlationId, "error", err)

		return nil, nil, err
	}

	return rsp, rspBody, nil
}

// httpPut makes a PUT request to the given URL and returns the response & response body.
func (h *HTTPClient) httpPut(ctx context.Context, //nolint:dupl
	url string, headers []Header, body any,
) (*http.Response, []byte, error) {
	req, err := makePutRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	correlationId := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		var serializedBody []byte

		if body != nil {
			serializedBody, err = json.Marshal(body)
			if err != nil {
				logging.Logger(ctx).Error("Failed to serialize request body",
					"method", "PUT", "url", url,
					"correlationId", correlationId, "error", err)
			}
		}

		logRequestWithBody(logging.VerboseLogger(ctx), req, "PUT", correlationId, url, serializedBody)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, "PUT", correlationId, url)
	}

	rsp, rspBody, err := h.sendRequest(req)

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, "PUT", correlationId, url, rspBody)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, "PUT", correlationId, url)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", "PUT", "url", url,
			"correlationId", correlationId, "error", err)

		return nil, nil, err
	}

	return rsp, rspBody, nil
}

// httpDelete makes a DELETE request to the given URL and returns the response & response body.
func (h *HTTPClient) httpDelete(ctx context.Context, //nolint:dupl
	url string, headers []Header,
) (*http.Response, []byte, error) {
	req, err := makeDeleteRequest(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}

	correlationId := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		logRequestWithoutBody(logging.VerboseLogger(ctx), req, "DELETE", correlationId, url)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, "DELETE", correlationId, url)
	}

	rsp, rspBody, err := h.sendRequest(req)

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, "DELETE", correlationId, url, rspBody)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, "DELETE", correlationId, url)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", "DELETE", "url", url,
			"correlationId", correlationId, "error", err)

		return nil, nil, err
	}

	return rsp, rspBody, nil
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
func makePostRequest(ctx context.Context, resourceURL string, headers Headers, data []byte) (*http.Request, error) {
	reader, _, err := bodyReader(headers, data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resourceURL, reader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return AddJSONContentTypeIfNotPresent(addHeaders(req, headers)), nil
}

// Determines how the payload should be provided to the HTTP request object based on the input headers.
//   - For "application/x-www-form-urlencoded", a map of strings to strings is expected.
//     The values are URL-encoded, and the new content length is calculated.
//   - For other content types, the payload data is sent as-is.
//
// Returns: An io.Reader for the request body, the content length, and an optional error.
func bodyReader(headers Headers, data []byte) (io.Reader, int64, error) {
	if headers.Has(HeaderFormURLEncoded) {
		var keyValuePairs map[string]string

		err := json.Unmarshal(data, &keyValuePairs)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: %w", ErrPayloadNotURLForm, err)
		}

		payloadValues := url.Values{}
		for k, v := range keyValuePairs {
			payloadValues.Set(k, v)
		}

		encodedPayload := payloadValues.Encode()
		encodedBytes := []byte(encodedPayload)

		return bytes.NewReader(encodedBytes), int64(len(encodedBytes)), nil
	}

	return bytes.NewReader(data), int64(len(data)), nil
}

// makePatchRequest creates a PATCH request with the given headers and body, and adds the
// Content-Type header. It then returns the request.
func makePatchRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return AddJSONContentTypeIfNotPresent(addHeaders(req, headers)), nil
}

// makePutRequest creates a PUT request with the given headers and body, and adds the
// Content-Type header. It then returns the request.
func makePutRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return AddJSONContentTypeIfNotPresent(addHeaders(req, headers)), nil
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

	shouldHandleError := h.ShouldHandleError
	if shouldHandleError == nil {
		// Default predicate: treat "non-2xx" responses as requiring error handling.
		shouldHandleError = func(response *http.Response) bool {
			return response.StatusCode < 200 || response.StatusCode > 299
		}
	}

	if shouldHandleError(res) {
		if h.ErrorHandler != nil {
			// Invoke the custom error handler.
			return res, body, h.ErrorHandler(res, body)
		}

		// Fallback to generic error interpretation.
		return res, body, InterpretError(res, body)
	}

	// Response may indicate a logical failure at the API level (e.g., a record-level error),
	// but it is not a fatal HTTP error. Connectors can handle it according to their contract.
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

// AddJSONContentTypeIfNotPresent adds the Content-Type header if it is not already present.
func AddJSONContentTypeIfNotPresent(req *http.Request) *http.Request {
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

func GetRequestHeaders(request *http.Request) Headers {
	if request == nil {
		return nil
	}

	// Pre-calculate the total number of header values
	totalValues := 0
	for _, values := range request.Header {
		totalValues += len(values)
	}

	// Corner case: if there are no headers, return nil
	if totalValues == 0 {
		return nil
	}

	// Initialize the headers slice with accurate capacity
	headers := make(Headers, 0, totalValues)

	// Populate the headers slice
	for key, values := range request.Header {
		for _, value := range values {
			headers = append(headers, Header{Key: key, Value: value})
		}
	}

	return headers
}

func GetResponseHeaders(response *http.Response) Headers {
	if response == nil {
		return nil
	}

	// Pre-calculate the total number of header values
	totalValues := 0
	for _, values := range response.Header {
		totalValues += len(values)
	}

	// Corner case: if there are no headers, return nil
	if totalValues == 0 {
		return nil
	}

	// Initialize the headers slice with accurate capacity
	headers := make(Headers, 0, totalValues)

	// Populate the headers slice
	for key, values := range response.Header {
		for _, value := range values {
			headers = append(headers, Header{Key: key, Value: value})
		}
	}

	return headers
}
