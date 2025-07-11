package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"github.com/spyzhov/ajson"
)

// JSONHTTPClient is an HTTP client which makes certain assumptions, such as
// that the response body is JSON. It also handles OAuth access token refreshes.
type JSONHTTPClient struct {
	HTTPClient         *HTTPClient        // underlying HTTP client. Required.
	ErrorPostProcessor ErrorPostProcessor // Errors returned from CRUD methods will go via this method. Optional.
}

// JSONHTTPResponse is a JSON response from an HTTP request.
// Consider using Body to operate on ajson.Node or
// Unmarshal into the struct of your choosing via UnmarshalJSON.
type JSONHTTPResponse struct {
	// bodyBytes is the raw response body. It's not JSON-unmarshalled.
	// We keep it around so that we can unmarshal it into a struct later,
	// if needed (via the UnmarshalJSON function).
	bodyBytes []byte

	// Code is the HTTP status code of the response.
	Code int

	// Headers are the HTTP headers of the response.
	Headers http.Header

	// body is the JSON-unmarshalled response body. Aside from the fact
	// that it's JSON-unmarshalled, it's identical to bodyBytes.
	// If there were no bytes this will be nil.
	body *ajson.Node
}

// Body returns JSON node. If it is empty the flag will indicate so.
// Empty response body is a special case and should be handled explicitly.
func (j *JSONHTTPResponse) Body() (*ajson.Node, bool) {
	if j.body == nil {
		return nil, false
	}

	return j.body, true
}

// Get makes a GET request to the given URL and returns the response body as a JSON object.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request. If errorHandler is nil, then the default error
// handler is used. If not, the caller can inject their own error handling logic.
func (j *JSONHTTPClient) Get(ctx context.Context, url string, headers ...Header) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Get(ctx, url, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return ParseJSONResponse(res, body)
}

// Post makes a POST request to the given URL and returns the response body as a JSON object.
// ReqBody must be JSON-serializable. If it is not, an error is returned.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request. If errorHandler is nil, then the default error
// handler is used. If not, the caller can inject their own error handling logic.
func (j *JSONHTTPClient) Post(ctx context.Context,
	url string, reqBody any, headers ...Header,
) (*JSONHTTPResponse, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", reqBody, err)
	}

	res, body, err := j.HTTPClient.Post(ctx, url, data, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return ParseJSONResponse(res, body)
}

func (j *JSONHTTPClient) Put(ctx context.Context,
	url string, reqBody any, headers ...Header,
) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Put(ctx, url, reqBody, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return ParseJSONResponse(res, body)
}

func (j *JSONHTTPClient) Patch(ctx context.Context,
	url string, reqBody any, headers ...Header,
) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Patch(ctx, url, reqBody, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return ParseJSONResponse(res, body)
}

func (j *JSONHTTPClient) Delete(ctx context.Context, url string, headers ...Header) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Delete(ctx, url, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return ParseJSONResponse(res, body)
}

// ParseJSONResponse parses the given HTTP response and returns a JSONHTTPResponse.
func ParseJSONResponse(res *http.Response, body []byte) (*JSONHTTPResponse, error) {
	// empty response body should not be parsed as JSON since it will cause ajson to err
	if len(body) == 0 {
		// Empty response. Both object and error are returned.
		// Caller must check for error.
		return &JSONHTTPResponse{
			bodyBytes: make([]byte, 0),
			Code:      res.StatusCode,
			Headers:   res.Header,
			body:      nil,
		}, nil
	}

	// Ensure the response is JSON.
	// Starts with application ends with JSON which may be followed by the version.
	if err := EnsureContentType(`^application/.*json([-0-9.])*$`, res, false); err != nil {
		return nil, err
	}

	// Unmarshall the response body into JSON
	jsonBody, err := ajson.Unmarshal(body)
	if err != nil {
		headers := GetResponseHeaders(res)

		return nil, NewHTTPError(res.StatusCode, body, headers,
			fmt.Errorf("failed to unmarshall response body into JSON: %w", err))
	}

	return &JSONHTTPResponse{
		bodyBytes: body,
		Code:      res.StatusCode,
		Headers:   res.Header,
		body:      jsonBody,
	}, nil
}

// UnmarshalJSON deserializes the response body into the given type.
func UnmarshalJSON[T any](rsp *JSONHTTPResponse) (*T, error) {
	var data T

	if len(rsp.bodyBytes) == 0 {
		// Empty struct.
		return &data, nil
	}

	if err := json.Unmarshal(rsp.bodyBytes, &data); err != nil {
		return nil, errors.Join(err, ErrFailedToUnmarshalBody)
	}

	return &data, nil
}

// MakeJSONGetRequest creates a GET request with the given headers and adds the
// Accept: application/json header. It then returns the request.
func MakeJSONGetRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	return MakeGetRequest(ctx, url, addAcceptJSONHeader(headers))
}

// addAcceptJSONHeader adds the Accept: application/json header to the given headers.
func addAcceptJSONHeader(headers []Header) []Header {
	if headers == nil {
		headers = make([]Header, 0)
	}

	return append(headers, Header{Key: "Accept", Value: "application/json"})
}

// AddSuffixIfNotExists appends the suffix  to the provided string.
func AddSuffixIfNotExists(str string, suffix string) string {
	if !strings.HasSuffix(str, suffix) {
		str += suffix
	}

	return str
}

var ErrMissingContentType = errors.New("missing content type")

// EnsureContentType ensures that the content type matches the given pattern.
// If errOnMissing is true, an error is returned if the content type is missing.
// Otherwise, the function returns nil.
func EnsureContentType(pattern string, res *http.Response, errOnMissing bool) error {
	ctype := res.Header.Get("Content-Type")

	// If the content type is missing, return an error if errOnMissing is true,
	// otherwise return nil.
	if len(ctype) == 0 {
		if errOnMissing {
			return fmt.Errorf("%w: expected content type to be application/*json", ErrMissingContentType)
		}

		return nil
	}

	mimeType, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return fmt.Errorf("failed to parse content type: %w", err)
	}

	// Check if the mimeType follows the pattern application/*json (e.g.
	// application/vnd.api+json, application/schema+json, etc.)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %w", err)
	}

	if !re.MatchString(mimeType) {
		return fmt.Errorf("%w: expected content type to be %s, got %s",
			ErrNotJSON, pattern, mimeType,
		)
	}

	return nil
}
