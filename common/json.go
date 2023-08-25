package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/spyzhov/ajson"
)

// ErrorHandler allows the caller to inject their own HTTP error handling logic.
// All non-2xx responses will be passed to the error handler. If the error handler
// returns nil, then the error is ignored and the caller is responsible for handling
// the error. If the error handler returns an error, then that error is returned
// to the caller, as-is. Both the response and the response body are passed
// to the error handler as arguments.
type ErrorHandler func(rsp *http.Response, body []byte) error

// JSONHTTPClient is an HTTP client which makes certain assumptions, such as
// that the response body is JSON. It also handles OAuth access token refreshes.
type JSONHTTPClient struct {
	Base         string                  // optional base URL. If not set, then all URLs must be absolute.
	Client       AuthenticatedHTTPClient // underlying HTTP client. Required.
	ErrorHandler ErrorHandler            // optional error handler. If not set, then the default error handler is used.
}

// Get makes a GET request to the given URL and returns the response body as a JSON object.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request. If errorHandler is nil, then the default error
// handler is used. If not, the caller can inject their own error handling logic.
func (j *JSONHTTPClient) Get(ctx context.Context, url string, headers ...Header) (*ajson.Node, error) {
	fullURL, err := j.getURL(url)
	if err != nil {
		return nil, err
	}

	// Make the request, get the response body
	res, body, err := j.httpGet(ctx, fullURL, headers) //nolint:bodyclose
	if err != nil {
		return nil, err
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
) (*ajson.Node, error) {
	fullURL, err := j.getURL(url)
	if err != nil {
		return nil, err
	}
	// Make the request, get the response body
	res, body, err := j.httpPost(ctx, fullURL, headers, reqBody) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return parseJSONResponse(res, body)
}

func (j *JSONHTTPClient) getURL(u string) (string, error) {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u, nil
	}

	if len(j.Base) == 0 {
		return "", fmt.Errorf("%w (input is %q)", ErrEmptyBaseURL, u)
	}

	return url.JoinPath(j.Base, u)
}

func parseJSONResponse(res *http.Response, body []byte) (*ajson.Node, error) {
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

func (j *JSONHTTPClient) httpGet(ctx context.Context, url string,
	headers []Header,
) (*http.Response, []byte, error) {
	req, err := makeJSONGetRequest(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}

	return j.sendRequest(req)
}

func (j *JSONHTTPClient) httpPost(ctx context.Context, url string,
	headers []Header, body any,
) (*http.Response, []byte, error) {
	req, err := makeJSONPostRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	return j.sendRequest(req)
}

func (j *JSONHTTPClient) sendRequest(req *http.Request) (*http.Response, []byte, error) {
	// Send the request
	res, err := j.Client.Do(req)
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
		if j.ErrorHandler != nil {
			return nil, nil, j.ErrorHandler(res, body)
		}

		return nil, nil, InterpretError(res, body)
	}

	return res, body, nil
}

func makeJSONGetRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return addAcceptJSONHeaders(req, headers)
}

func makeJSONPostRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
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

	return addAcceptJSONHeaders(req, headers)
}

func addAcceptJSONHeaders(req *http.Request, headers []Header) (*http.Request, error) {
	// Request JSON
	req.Header.Add("Accept", "application/json")

	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req, nil
}
