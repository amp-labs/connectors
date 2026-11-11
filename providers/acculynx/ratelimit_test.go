package acculynx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

// throttledServer answers the first `throttled` requests with a 429 carrying
// the given Retry-After, then hands every later request to next. It counts
// every request it receives.
func throttledServer(t *testing.T, throttled int32, retryAfter string, next http.HandlerFunc) (
	*httptest.Server, *atomic.Int32,
) {
	t.Helper()

	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) <= throttled {
			if retryAfter != "" {
				w.Header().Set("Retry-After", retryAfter)
			}

			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("Too Many Requests"))

			return
		}

		next(w, r)
	}))
	t.Cleanup(srv.Close)

	return srv, &hits
}

func okHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

func doRequest(t *testing.T, ctx context.Context, client common.AuthenticatedHTTPClient, method, url string) (
	*http.Response, error,
) {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	assert.NilError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	return resp, err
}

func TestRetryOn429_RetriesGetUntilItSucceeds(t *testing.T) {
	t.Parallel()

	srv, hits := throttledServer(t, 2, "0", okHandler(`{}`))

	resp, err := doRequest(t, context.Background(), newRetryOn429Client(&http.Client{}), http.MethodGet, srv.URL)
	assert.NilError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	assert.Equal(t, hits.Load(), int32(3))
}

func TestRetryOn429_GivesUpAfterMaxRetries(t *testing.T) {
	t.Parallel()

	srv, hits := throttledServer(t, 100, "0", okHandler(`{}`))

	resp, err := doRequest(t, context.Background(), newRetryOn429Client(&http.Client{}), http.MethodGet, srv.URL)
	assert.NilError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusTooManyRequests)
	assert.Equal(t, hits.Load(), int32(maxRetriesOn429+1))
}

func TestRetryOn429_DoesNotRetryWrites(t *testing.T) {
	t.Parallel()

	srv, hits := throttledServer(t, 1, "0", okHandler(`{}`))

	resp, err := doRequest(t, context.Background(), newRetryOn429Client(&http.Client{}), http.MethodPost, srv.URL)
	assert.NilError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusTooManyRequests)
	assert.Equal(t, hits.Load(), int32(1))
}

func TestRetryOn429_StopsWaitingWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	srv, hits := throttledServer(t, 1, "60", okHandler(`{}`))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	started := time.Now()
	_, err := doRequest(t, ctx, newRetryOn429Client(&http.Client{}), http.MethodGet, srv.URL)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Assert(t, time.Since(started) < 5*time.Second, "waited the full Retry-After despite cancellation")
	assert.Equal(t, hits.Load(), int32(1))
}

func TestRetryDelayOn429(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		retryAfter string
		wantDelay  time.Duration
		wantRetry  bool
	}{
		{name: "uses Retry-After seconds", retryAfter: "60", wantDelay: 60 * time.Second, wantRetry: true},
		{name: "zero means retry now", retryAfter: "0", wantDelay: 0, wantRetry: true},
		{name: "missing header falls back to default", retryAfter: "", wantDelay: defaultRetryDelayOn429, wantRetry: true},
		{name: "unparsable header falls back to default", retryAfter: "soon", wantDelay: defaultRetryDelayOn429, wantRetry: true},
		{name: "too long a wait is not retried", retryAfter: "3600", wantRetry: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			if tt.retryAfter != "" {
				header.Set("Retry-After", tt.retryAfter)
			}

			delay, retry := retryDelayOn429(header)
			assert.Equal(t, retry, tt.wantRetry)

			if tt.wantRetry {
				assert.Equal(t, delay, tt.wantDelay)
			}
		})
	}
}

// The per-contact fetch that failed in production must survive a 429.
func TestGetRecordsByIds_RetriesAfter429(t *testing.T) {
	t.Parallel()

	srv, hits := throttledServer(t, 1, "0", okHandler(`{"id":"ctc-1"}`))

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	result, err := conn.GetRecordsByIds(context.Background(), objectContacts, []string{"ctc-1"}, []string{"id"}, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(result.Rows), 1)
	assert.Equal(t, hits.Load(), int32(2))
}

// The list reader holds its own copy of the HTTP client, so it needs the retry too.
func TestRead_RetriesListCallAfter429(t *testing.T) {
	t.Parallel()

	srv, hits := throttledServer(t, 1, "0", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v2/company-settings/custom-fields" {
			_, _ = w.Write(customFieldDefinitionsEmptyResponse)

			return
		}

		_, _ = w.Write(contactsIncludesResponse)
	})

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	result, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: objectContacts,
		Fields:     connectors.Fields("id"),
	})
	assert.NilError(t, err)
	assert.Equal(t, result.Rows, int64(1))
	assert.Assert(t, hits.Load() >= 2)
}
