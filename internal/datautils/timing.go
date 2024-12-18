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

// FormatRFC3339inUTCWithMilliseconds similar to FormatRFC3339inUTC but adds milliseconds.
// 15:04:05.039 => milliseconds is a decimal part of seconds.
func (timing) FormatRFC3339inUTCWithMilliseconds(input time.Time) string {
	utcTime := input.UTC()

	return utcTime.Format("2006-01-02T15:04:05.000Z")
}
