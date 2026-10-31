package twenty

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// What an error becomes decides what the server does: retry later, stop, or refresh the token.
func TestErrorClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		file       string
		body       string
		mime       string
		headers    map[string]string
		wantErr    error
		wantClass  common.ErrorClass
		wantHeader string
		retryable  bool
	}{
		{
			name:   "429 is a rate limit, with status, body and Retry-After kept for the server",
			status: http.StatusTooManyRequests,
			file:   "error-rate-limited.json",
			headers: map[string]string{
				"Retry-After":           "1",
				"X-RateLimit-Limit":     "100",
				"X-RateLimit-Remaining": "0",
			},
			wantErr:    common.ErrLimitExceeded,
			wantClass:  common.ErrorClassRateLimited,
			wantHeader: "Retry-After",
		},
		{
			name:      "404 for a missing record is not retried",
			status:    http.StatusNotFound,
			file:      "error-record-not-found.json",
			wantErr:   common.ErrNotFound,
			wantClass: common.ErrorClassBadRequest,
		},
		{
			name:      "404 page in HTML (wrong domain, a proxy in front) is not retried",
			status:    http.StatusNotFound,
			body:      "<html><body><h1>404 Not Found</h1></body></html>",
			mime:      "text/html",
			wantErr:   common.ErrNotFound,
			wantClass: common.ErrorClassBadRequest,
		},
		{
			name:      "404 without a Content-Type is not retried",
			status:    http.StatusNotFound,
			body:      "Not Found",
			mime:      "none",
			wantErr:   common.ErrNotFound,
			wantClass: common.ErrorClassBadRequest,
		},
		{
			name:      "400 for an unknown object is not retried",
			status:    http.StatusBadRequest,
			file:      "error-object-not-found.json",
			wantErr:   common.ErrBadRequest,
			wantClass: common.ErrorClassBadRequest,
		},
		{
			name:      "408 from a proxy in front of the instance is retried",
			status:    http.StatusRequestTimeout,
			body:      "<html><body><h1>408 Request Time-out</h1></body></html>",
			mime:      "text/html",
			wantErr:   common.ErrRetryable,
			wantClass: common.ErrorClassRetryable,
			retryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body := []byte(tt.body)
			if tt.file != "" {
				body = testutils.DataFromFile(t, tt.file)
			}

			setup := []http.HandlerFunc{mockserver.ContentJSON()}

			switch tt.mime {
			case "":
			case "none":
				setup = []http.HandlerFunc{func(w http.ResponseWriter, _ *http.Request) { w.Header()["Content-Type"] = nil }}
			default:
				setup = []http.HandlerFunc{mockserver.ContentMIME(tt.mime)}
			}
			for key, value := range tt.headers {
				setup = append(setup, mockserver.Header(key, value))
			}

			server := mockserver.Fixed{
				Setup:  mockserver.ResponseChainedFuncs(setup...),
				Always: mockserver.Response(tt.status, body),
			}.Server()
			defer server.Close()

			conn, err := constructTestConnector(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			_, err = conn.Read(t.Context(), common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("id")})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error %v, want %v", err, tt.wantErr)
			}

			if errors.Is(err, common.ErrRetryable) != tt.retryable {
				t.Errorf("error %v: ErrRetryable is %v, want %v", err, !tt.retryable, tt.retryable)
			}

			if class := common.ClassOf(err); class != tt.wantClass {
				t.Errorf("class %q, want %q", class, tt.wantClass)
			}

			var httpErr *common.HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("error %v does not carry the HTTP response", err)
			}

			if httpErr.Status != tt.status || string(httpErr.Body) != string(body) {
				t.Errorf("kept status %d and body %q, want %d and the provider body", httpErr.Status, httpErr.Body, tt.status)
			}

			if tt.wantHeader != "" && !hasHeader(httpErr.Headers, tt.wantHeader, tt.headers[tt.wantHeader]) {
				t.Errorf("header %s=%s not kept: %v", tt.wantHeader, tt.headers[tt.wantHeader], httpErr.Headers)
			}
		})
	}
}

func hasHeader(headers common.Headers, key, value string) bool {
	for _, header := range headers {
		if http.CanonicalHeaderKey(header.Key) == http.CanonicalHeaderKey(key) && header.Value == value {
			return true
		}
	}

	return false
}
