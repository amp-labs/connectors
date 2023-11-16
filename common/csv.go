package common

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

func (c *HTTPClient) PutCSV(ctx context.Context, url string, reqBody []byte, headers ...Header) ([]byte, error) {
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

func (c *HTTPClient) GetCSV(ctx context.Context, url string, headers ...Header) ([]byte, error) {
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

func (c *HTTPClient) httpPutCSV(ctx context.Context, url string,
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
	// Request JSON
	req.Header.Add("Accept", "csv/text")

	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req, nil
}
