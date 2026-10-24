package circleback

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Circleback limits requests per account in two windows, per second and per minute, whose sizes
// depend on the plan: 3/s and 20/min on Free, 20/s and 300/min on Pro, Business and Enterprise.
// Every response reports the window closest to exhaustion in RateLimit-Remaining and
// RateLimit-Reset (seconds until it refills), and a 429 carries Retry-After (seconds).
// Header names are in Go's canonical form; the docs write them RateLimit-Remaining and RateLimit-Reset.
// https://circleback.ai/docs/api#rate-limits
const (
	headerRateLimitRemaining = "Ratelimit-Remaining"
	headerRateLimitReset     = "Ratelimit-Reset"
	headerRetryAfter         = "Retry-After"
)

// pacer holds requests back until the account's rate-limit window refills. It learns from the
// headers of every response, so it paces to whichever window and plan the account is on.
type pacer struct {
	mutex     sync.Mutex
	notBefore time.Time
}

// wait blocks until requests may be sent again.
func (p *pacer) wait(ctx context.Context) error {
	p.mutex.Lock()
	until := p.notBefore
	p.mutex.Unlock()

	delay := time.Until(until)
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// observe holds requests back for Retry-After, or for RateLimit-Reset once RateLimit-Remaining is 0.
func (p *pacer) observe(headers http.Header) {
	if seconds, ok := headerSeconds(headers, headerRetryAfter); ok {
		p.holdFor(seconds)

		return
	}

	if headers.Get(headerRateLimitRemaining) != "0" {
		return
	}

	if seconds, ok := headerSeconds(headers, headerRateLimitReset); ok {
		p.holdFor(seconds)
	}
}

func (p *pacer) holdFor(seconds int) {
	until := time.Now().Add(time.Duration(seconds) * time.Second)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if until.After(p.notBefore) {
		p.notBefore = until
	}
}

func headerSeconds(headers http.Header, name string) (int, bool) {
	seconds, err := strconv.Atoi(headers.Get(name))
	if err != nil || seconds < 0 {
		return 0, false
	}

	return seconds, true
}
