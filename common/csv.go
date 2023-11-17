package common

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// JSONHTTPClient is an HTTP client which makes certain assumptions, such as
// that the response body is JSON. It also handles OAuth access token refreshes.
type CSVHTTPClient struct {
	Base         string                  // optional base URL. If not set, then all URLs must be absolute.
	Client       AuthenticatedHTTPClient // underlying HTTP client. Required.
	ErrorHandler ErrorHandler            // optional error handler. If not set, then the default error handler is used.
}

func (c *CSVHTTPClient) getURL(url string) (string, error) {
	return getURL(c.Base, url)
}

func (c *CSVHTTPClient) PutCSV(ctx context.Context, url string, reqBody []byte, headers ...Header) ([]byte, error) {
	fullURL, err := c.getURL(url)
	if err != nil {
		return nil, err
	}

	_, body, err := c.httpPutCSV(ctx, fullURL, headers, reqBody) // nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *CSVHTTPClient) GetCSV(ctx context.Context, url string, headers ...Header) ([]byte, error) {
	fullURL, err := c.getURL(url)
	if err != nil {
		return nil, err
	}

	// Make the request, get the response body
	_, body, err := c.httpGet(ctx, fullURL, headers) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("error in httpGet: %w", err)
	}

	return body, nil
}

func (c *CSVHTTPClient) httpPutCSV(ctx context.Context, url string,
	headers []Header, body []byte,
) (*http.Response, []byte, error) {
	req, err := makeTextCSVPutRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	return c.sendRequest(req)
}

func makeTextCSVPutRequest(ctx context.Context, url string, headers []Header, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "text/csv"})
	req.ContentLength = int64(len(body))

	return addHeaders(req, headers)
}

func addHeaders(req *http.Request, headers []Header) (*http.Request, error) {
	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req, nil
}
