package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
)

/*
  // TODO: Below is work around to have CSV file type to be supported
// Currently done on JSONHTTPClient, but we might need a separate CSVHTTPClient instead
  // Research and see if there is a better way to do this
*/

// ErrMissingCSVData is returned when no CSV data was given for the upload.
var ErrMissingCSVData = errors.New("no CSV data provided")

func (j *JSONHTTPClient) PutCSV(ctx context.Context, url string, reqBody []byte, headers ...Header) ([]byte, error) {
	fullURL, err := j.HTTPClient.getURL(url)
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	_, body, err := j.httpPutCSV(ctx, fullURL, headers, reqBody) // nolint:bodyclose
	if err != nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	return body, nil
}

func (j *JSONHTTPClient) httpPutCSV(ctx context.Context, url string,
	headers []Header, body []byte,
) (*http.Response, []byte, error) {
	req, err := makeTextCSVPutRequest(ctx, url, headers, body)
	if err != nil {
		return nil, nil, err
	}

	return j.HTTPClient.sendRequest(req)
}

func makeTextCSVPutRequest(ctx context.Context, url string, headers []Header, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "text/csv"})
	req.ContentLength = int64(len(body))

	return addHeaders(req, headers), nil
}

// TODO: to be migrated to CSVHTTPClient once implemented
// func (j *JSONHTTPClient) GetCSV(ctx context.Context, url string, headers ...Header) ([]byte, error) {
// 	fullURL, err := j.getURL(url)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Make the request, get the response body
// 	_, body, err := j.httpGet(ctx, fullURL, headers) //nolint:bodyclose
// 	if err != nil {
// 		return nil, fmt.Errorf("error in httpGet: %w", err)
// 	}

// 	return body, nil
// }
