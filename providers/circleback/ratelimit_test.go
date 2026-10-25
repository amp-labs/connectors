package circleback

import (
	"net/http"
	"testing"
	"time"
)

func TestPacerObserve(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name    string
		headers http.Header
		hold    time.Duration
	}{
		{
			name:    "Requests left in the window: no hold",
			headers: http.Header{"Ratelimit-Remaining": {"14"}, "Ratelimit-Reset": {"24"}},
			hold:    0,
		},
		{
			name:    "Window used up: hold until it refills",
			headers: http.Header{"Ratelimit-Remaining": {"0"}, "Ratelimit-Reset": {"24"}},
			hold:    24 * time.Second,
		},
		{
			name:    "Rate limited: hold for Retry-After",
			headers: http.Header{"Retry-After": {"7"}, "Ratelimit-Remaining": {"0"}, "Ratelimit-Reset": {"30"}},
			hold:    7 * time.Second,
		},
		{
			name:    "No rate-limit headers: no hold",
			headers: http.Header{},
			hold:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var limiter pacer

			before := time.Now()
			limiter.observe(tt.headers)

			if tt.hold == 0 {
				if !limiter.notBefore.IsZero() {
					t.Fatalf("expected no hold, got one until %v", limiter.notBefore)
				}

				return
			}

			held := limiter.notBefore.Sub(before)
			if held < tt.hold || held > tt.hold+time.Second {
				t.Fatalf("expected a hold of %v, got %v", tt.hold, held)
			}
		})
	}
}
