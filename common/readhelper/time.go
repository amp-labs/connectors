package readhelper

import (
	"time"

	"github.com/amp-labs/connectors/common"
)

// TimeOrder describes the chronological ordering of records within a response.
type TimeOrder int

const (
	// Unordered means no assumptions can be made about read record ordering.
	Unordered TimeOrder = iota
	// ChronologicalOrder means records are ordered from oldest to newest.
	ChronologicalOrder
	// ReverseOrder means records are ordered from newest to oldest.
	ReverseOrder
)

// TimeBoundary controls the inclusivity or exclusivity of time-based filters
// (Since/Until) when evaluating whether a record timestamp falls within range.
type TimeBoundary struct {
	excludeSince bool
	excludeUntil bool
}

// TimeBoundaryOption is a function that modifies a TimeBoundary.
type TimeBoundaryOption func(*TimeBoundary)

// ExcludeSince makes the Since boundary exclusive.
func ExcludeSince() TimeBoundaryOption {
	return func(b *TimeBoundary) { b.excludeSince = true }
}

// ExcludeUntil makes the Until boundary exclusive.
func ExcludeUntil() TimeBoundaryOption {
	return func(b *TimeBoundary) { b.excludeUntil = true }
}

// NewTimeBoundary constructs a TimeBoundary using the provided options.
//
// By default, both Since and Until are inclusive.
//
// Examples:
//
//	b := NewTimeBoundary() 									// since <= timestamp <= until
//	b := NewTimeBoundary(ExcludeSince())					// since < timestamp <= until
//	b := NewTimeBoundary(ExcludeUntil())					// since <= timestamp < until
//	b := NewTimeBoundary(ExcludeSince(), ExcludeUntil())	// since < timestamp < until
func NewTimeBoundary(opts ...TimeBoundaryOption) *TimeBoundary {
	b := &TimeBoundary{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Contains checks whether a given timestamp falls within the Since/Until range
// defined in params, respecting the current TimeBoundary mode.
//
// If both Since and Until are zero, Contains always returns true.
func (b TimeBoundary) Contains(params common.ReadParams, timestamp time.Time) bool {
	// If neither Since nor Until is provided, always allow
	if params.Since.IsZero() && params.Until.IsZero() {
		return true
	}

	sinceOK := true
	untilOK := true

	// Check Since only if it's non-zero.
	// Otherwise, allow.
	if !params.Since.IsZero() {
		if b.excludeSince {
			// Strict comparison not including since timestamp.
			sinceOK = timestamp.After(params.Since)
		} else {
			// Everything after since, including since.
			sinceOK = !timestamp.Before(params.Since)
		}
	}

	// Check Until only if it's non-zero.
	// Otherwise, allow.
	if !params.Until.IsZero() {
		if b.excludeUntil {
			// Strict comparison not including until timestamp.
			untilOK = timestamp.Before(params.Until)
		} else {
			// Everything before until, including until.
			untilOK = !timestamp.After(params.Until)
		}
	}

	return sinceOK && untilOK
}

func (b TimeBoundary) Before(params common.ReadParams, timestamp time.Time) bool {
	if params.Since.IsZero() {
		return false // boundary goes to infinity (extends to the past).
	}

	if b.excludeSince && timestamp.Equal(params.Since) {
		return false
	}

	return timestamp.Before(params.Since)
}

func (b TimeBoundary) After(params common.ReadParams, timestamp time.Time) bool {
	if params.Until.IsZero() {
		return false // boundary goes to infinity (extends to the future).
	}

	if b.excludeUntil && timestamp.Equal(params.Until) {
		return false
	}

	return timestamp.After(params.Until)
}
