package datautils

import (
	"log/slog"
	"net/http"
)

var HTTP = httpUtils{} // nolint:gochecknoglobals

type httpUtils struct{}

func (httpUtils) BodyClose(response *http.Response) func() {
	return func() {
		if response != nil && response.Body != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}
}

func (httpUtils) IsStatus2XX(response *http.Response) bool {
	return 200 <= response.StatusCode && response.StatusCode < 300
}
