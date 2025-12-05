// nolint:revive
package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/logging"
	"github.com/google/uuid"
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

	correlationId := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		logRequestWithBody(logging.VerboseLogger(ctx), req, "PUT", correlationId, url, body)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, "PUT", correlationId, url)
	}

	rsp, body, err := j.HTTPClient.sendRequest(req)

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, "PUT", correlationId, url, body)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, "PUT", correlationId, url)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", "PUT", "url", url,
			"correlationId", correlationId, "error", err)

		return nil, nil, err
	}

	return rsp, body, nil
}

func makeTextCSVPutRequest(ctx context.Context, url string, headers []Header, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "text/csv"})

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
