package datautils

import (
	"strconv"
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

// Unix returns the string representing the number of seconds elapsed since January 1, 1970.
func (timing) Unix(input time.Time) string {
	return strconv.FormatInt(input.Unix(), 10)
}

// FormatRFC3339WithOffset is RFC3339 but includes time zone offset.
func (timing) FormatRFC3339WithOffset(input time.Time) string {
	return input.Format("2006-01-02T15:04:05-07:00")
}
