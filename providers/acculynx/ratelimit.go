package acculynx

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
)

// AccuLynx limits each API key to 10 requests per second and each IP to 30 requests
// per second (https://apidocs.acculynx.com/docs/rate-limits). Our egress IPs are shared
// by every AccuLynx customer, so a 429 can arrive even while one read stays within
// its own key's limit. Each 429 carries Retry-After (60 seconds in production).
//
// Without a retry, one 429 on any of a page's many child fetches fails the whole page,
// and the caller only tries again after its own backoff, which grows to an hour.
// Waiting out Retry-After and resending that one request keeps the read moving.
const (
	// maxRetriesOn429 is how many times one request is resent after a 429.
	maxRetriesOn429 = 3

	// defaultRetryDelayOn429 is used when a 429 has no usable Retry-After.
	defaultRetryDelayOn429 = 60 * time.Second

	// maxRetryDelayOn429 is the longest we wait in-process. A longer Retry-After
	// (e.g. a write quota reset) is returned to the caller instead.
	maxRetryDelayOn429 = 2 * time.Minute
)

// retryOn429Client resends a GET that AccuLynx answered with 429, after waiting the
// Retry-After delay. Only GETs are retried: they have no body to replay and reads are
// where the 429s happen. Everything else passes straight through.
type retryOn429Client struct {
	common.AuthenticatedHTTPClient
}

func newRetryOn429Client(client common.AuthenticatedHTTPClient) *retryOn429Client {
	return &retryOn429Client{AuthenticatedHTTPClient: client}
}

func (c *retryOn429Client) Do(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := c.AuthenticatedHTTPClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests ||
			req.Method != http.MethodGet || attempt == maxRetriesOn429 {
			return resp, err
		}

		delay, ok := retryDelayOn429(resp.Header)
		if !ok {
			return resp, nil
		}

		// Free the connection before waiting; the body of a 429 is not needed.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		logging.Logger(req.Context()).Warn("AccuLynx rate limited a request, retrying after delay",
			"path", req.URL.Path, "retry", attempt+1, "delay", delay)

		if err := sleep(req, delay); err != nil {
			return nil, err
		}
	}
}

// retryDelayOn429 returns how long to wait before resending, and false when the
// wait is too long to do in-process.
func retryDelayOn429(header http.Header) (time.Duration, bool) {
	delay := defaultRetryDelayOn429

	if seconds, err := strconv.Atoi(header.Get("Retry-After")); err == nil && seconds >= 0 {
		delay = time.Duration(seconds) * time.Second
	}

	return delay, delay <= maxRetryDelayOn429
}

// sleep waits for delay, or returns early with the request's context error.
func sleep(req *http.Request, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-req.Context().Done():
		return req.Context().Err()
	}
}
