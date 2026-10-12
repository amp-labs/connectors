package mail

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

// SearchBoxQuery represents a Gmail-compatible filter for mail search queries.
// It constructs a `q` parameter using `after:` and `before:` with YYYY/MM/DD format or Unix time.
//
// This is intended for incremental reads of Gmail collection endpoints (messages, drafts, threads).
type SearchBoxQuery struct {
	since      string
	until      string
	conditions []string
}

func newSearchBoxQuery() SearchBoxQuery {
	return SearchBoxQuery{}
}

func (q SearchBoxQuery) WithCondition(condition string) SearchBoxQuery {
	q.conditions = append(q.conditions, condition)

	return q
}

func (q SearchBoxQuery) WithSince(timestamp time.Time) SearchBoxQuery {
	if !timestamp.IsZero() {
		q.since = datautils.Time.Unix(timestamp)
	}

	return q
}

func (q SearchBoxQuery) WithUntil(timestamp time.Time) SearchBoxQuery {
	if !timestamp.IsZero() {
		q.until = datautils.Time.Unix(timestamp)
	}

	return q
}

func (q SearchBoxQuery) String() string {
	var conditions []string

	if q.since != "" {
		conditions = append(conditions, fmt.Sprintf("after:%v", q.since))
	}

	if q.until != "" {
		conditions = append(conditions, fmt.Sprintf("before:%v", q.until))
	}

	if len(q.conditions) != 0 {
		conditions = append(conditions, q.conditions...)
	}

	return strings.Join(conditions, " ")
}
