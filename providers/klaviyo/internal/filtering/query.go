// Package filtering provides a helper to construct filtering query parameters
// compatible with APIs that support time-based and custom filtering syntax,
// such as Klaviyo's ?filter=... convention.
package filtering

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

// Query represents a composed filter string that can include
// time-based and custom filters for a request.
type Query struct {
	since  string
	until  string
	custom string
}

// NewQuery initializes an empty filtering query which can be expanded using builder methods.
func NewQuery() *Query {
	return &Query{}
}

// WithSince adds a filter to retrieve records updated after the given timestamp.
// If the timestamp is zero, the method is a no-op.
//
// Example output (Klaviyo format):
//
//	`greater-than(updated_at,2023-03-01T01:00:00Z)`
//
// See: https://developers.klaviyo.com/en/docs/filtering_
func (q *Query) WithSince(timestamp time.Time, fieldName string) *Query {
	if timestamp.IsZero() {
		// No-op
		return q
	}

	timeValue := datautils.Time.FormatRFC3339inUTC(timestamp)
	q.since = fmt.Sprintf("greater-than(%v,%v)", fieldName, timeValue)

	return q
}

// WithUntil adds a filter to retrieve records updated before the given timestamp.
// If the timestamp is zero, the method is a no-op.
//
// Example output:
//
//	`less-than(updated_at,2023-03-01T00:00:00Z)`
//
// See: https://developers.klaviyo.com/en/docs/filtering_
func (q *Query) WithUntil(timestamp time.Time, fieldName string) *Query {
	if timestamp.IsZero() {
		// No-op
		return q
	}

	timeValue := datautils.Time.FormatRFC3339inUTC(timestamp)
	q.until = fmt.Sprintf("less-than(%v,%v)", fieldName, timeValue)

	return q
}

// WithCustomFiltering adds arbitrary filters to the query.
// This can be used for provider-specific conditions outside time-based logic.
//
// Filters must be formatted as a comma-separated string, e.g.:
//
//	`equals(email,"sarah.mason@klaviyo-demo.com")`
//	`contains(name,"marketing")`
//
// See: https://developers.klaviyo.com/en/docs/filtering_
func (q *Query) WithCustomFiltering(customFilters string) *Query {
	q.custom = customFilters

	return q
}

// String returns the final filter string, combining since/until/custom parts.
// The filters are joined using commas, which Klaviyo APIs interpret as implicit logical AND.
//
// See: https://developers.klaviyo.com/en/docs/filtering_#boolean-logic-operators
func (q *Query) String() string {
	filters := make([]string, 0)

	if q.since != "" {
		filters = append(filters, q.since)
	}

	if q.until != "" {
		filters = append(filters, q.until)
	}

	if q.custom != "" {
		filters = append(filters, q.custom)
	}

	// As per documentation filtering values are comma separated.
	// Reference: https://developers.klaviyo.com/en/docs/filtering_
	return strings.Join(filters, ",")
}
