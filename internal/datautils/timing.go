package datautils

import (
	"time"
)

type timing struct{}

var Time timing // nolint:gochecknoglobals

// FormatRFC3339inUTC will convert time to UTC respecting the time zone difference
// and only then apply time.RFC3339 format.
func (timing) FormatRFC3339inUTC(input time.Time) string {
	// Format cuts the zone, nanoseconds and other parts which are not used.
	// Time zone should be preserved.
	// Apply zone, keeping time difference.
	utcTime := input.UTC()

	return utcTime.Format("2006-01-02T15:04:05Z")
}
