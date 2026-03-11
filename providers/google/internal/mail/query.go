package mail

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

// TimeQuery represents a Gmail-compatible time filter for search queries.
// It constructs a `q` parameter using `after:` and `before:` with YYYY/MM/DD format or Unix time.
//
// This is intended for incremental reads of Gmail collection endpoints (messages, drafts, threads).
type TimeQuery struct {
	since string
	until string
}

func newTimeQuery() *TimeQuery {
	return &TimeQuery{}
}

func (q *TimeQuery) WithSince(timestamp time.Time) *TimeQuery {
	if !timestamp.IsZero() {
		q.since = datautils.Time.Unix(timestamp)
	}

	return q
}

func (q *TimeQuery) WithUntil(timestamp time.Time) *TimeQuery {
	if !timestamp.IsZero() {
		q.until = datautils.Time.Unix(timestamp)
	}

	return q
}

func (q *TimeQuery) String() string {
	if q.since == "" && q.until == "" {
		return ""
	}

	if q.since != "" && q.until != "" {
		return fmt.Sprintf("after:%v before:%v", q.since, q.until)
	}

	if q.since != "" {
		return fmt.Sprintf("after:%v", q.since)
	}

	return fmt.Sprintf("before:%v", q.until)
}
