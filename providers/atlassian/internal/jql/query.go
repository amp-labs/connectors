package jql

import (
	"fmt"
	"time"
)

// JQL is Jira Query Language.
//
// Read URL supports time scoping. common.ReadParams.Since is used to get relative time frame.
// Here is an API example on how to request issues that were updated in the last 30 minutes.
// search?jql=updated > "-30m"
// The reason we use minutes is that it is the most granular API permits.
type JQL struct {
	since string
	until string
}

func New() *JQL {
	return new(JQL)
}

func (q *JQL) SinceMinutes(since time.Time) *JQL {
	q.since = relativeMinutesFromNow(since)

	return q
}

func (q *JQL) UntilMinutes(until time.Time) *JQL {
	q.until = relativeMinutesFromNow(until)

	return q
}

func (q *JQL) String() string {
	// Between. Both is specified.
	if q.since != "" && q.until != "" {
		return fmt.Sprintf(`updated > "%vm" AND updated < "%vm"`, q.since, q.until)
	}

	if q.since != "" {
		return fmt.Sprintf(`updated > "%vm"`, q.since)
	}

	if q.until != "" {
		return fmt.Sprintf(`updated < "%vm"`, q.until)
	}

	// Neither since nor until is specified. Request records updated before now, which is all records.
	return `updated < "0m"`
}
