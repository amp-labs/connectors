package common

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/spyzhov/ajson"
)

// JSONHTTPClient is an HTTP client which makes certain assumptions, such as
// that the response body is JSON. It also handles OAuth access token refreshes.
type JSONHTTPClient struct {
	HTTPClient         *HTTPClient        // underlying HTTP client. Required.
	ErrorPostProcessor ErrorPostProcessor // Errors returned from CRUD methods will go via this method. Optional.
}

// JSONHTTPResponse is a JSON response from an HTTP request.
type JSONHTTPResponse struct {
	// bodyBytes is the raw response body. It's not JSON-unmarshalled.
	// We keep it around so that we can unmarshal it into a struct later,
	// if needed (via the UnmarshalJSON function).
	bodyBytes []byte

	// Code is the HTTP status code of the response.
	Code int

	// Headers are the HTTP headers of the response.
	Headers http.Header

	// Body is the JSON-unmarshalled response body. Aside from the fact
	// that it's JSON-unmarshalled, it's identical to bodyBytes.
	Body *ajson.Node
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

	return parseJSONResponse(res, body)
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

	return parseJSONResponse(res, body)
}

func (j *JSONHTTPClient) Put(ctx context.Context,
	url string, reqBody any, headers ...Header,
) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Put(ctx, url, reqBody, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return parseJSONResponse(res, body)
}

func (j *JSONHTTPClient) Patch(ctx context.Context,
	url string, reqBody any, headers ...Header,
) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Patch(ctx, url, reqBody, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return parseJSONResponse(res, body)
}

func (j *JSONHTTPClient) Delete(ctx context.Context, url string, headers ...Header) (*JSONHTTPResponse, error) {
	res, body, err := j.HTTPClient.Delete(ctx, url, addAcceptJSONHeader(headers)...) //nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return parseJSONResponse(res, body)
}

// parseJSONResponse parses the given HTTP response and returns a JSONHTTPResponse.
func parseJSONResponse(res *http.Response, body []byte) (*JSONHTTPResponse, error) {
	// empty response body should not be parsed as JSON since it will cause ajson to err
	if len(body) == 0 {
		return nil, nil //nolint:nilnil
	}
	// Ensure the response is JSON
	ct := res.Header.Get("Content-Type")
	if len(ct) > 0 {
		mimeType, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content type: %w", err)
		}

		// Providers implementing JSONAPISpeicifcations returns application/vnd.api+json
		if mimeType != "application/json" && mimeType != "application/vnd.api+json" {
			return nil, fmt.Errorf("%w: expected content type to be application/json or application/vnd.api+json , got %s",
				ErrNotJSON, mimeType,
			)
		}
	}

	// Unmarshall the response body into JSON
	jsonBody, err := ajson.Unmarshal(body)
	if err != nil {
		return nil, NewHTTPStatusError(res.StatusCode, fmt.Errorf("failed to unmarshall response body into JSON: %w", err))
	}

	return &JSONHTTPResponse{
		bodyBytes: body,
		Code:      res.StatusCode,
		Headers:   res.Header,
		Body:      jsonBody,
	}, nil
}

// UnmarshalJSON deserializes the response body into the given type.
func UnmarshalJSON[T any](rsp *JSONHTTPResponse) (*T, error) {
	var data T

	if err := json.Unmarshal(rsp.bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body into JSON: %w", err)
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
