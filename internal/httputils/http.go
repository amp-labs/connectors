package httputils

import (
	"log/slog"
	"net/http"
)

func BodyClose(response *http.Response) func() {
	return func() {
		if response != nil && response.Body != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}
}

func IsStatus2XX(response *http.Response) bool {
	return 200 <= response.StatusCode && response.StatusCode < 300
}
