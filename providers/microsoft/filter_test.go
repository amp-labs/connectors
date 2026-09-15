package microsoft

import (
	"testing"
	"time"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestFilterQuery_String(t *testing.T) {
	t.Parallel()

	sinceTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	untilTime := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)

	// Since we are testing internal implementation of since/until formatting,
	// we need to know how it formats the time.
	// filter.go uses datautils.Time.FormatRFC3339inUTCWithMilliseconds(timestamp)
	sinceStr := "2023-01-01T12:00:00.000Z"
	untilStr := "2023-01-02T12:00:00.000Z"

	tests := []struct {
		name     string
		setup    func(q filterQuery) filterQuery
		expected string
	}{
		{
			name: "Since only",
			setup: func(q filterQuery) filterQuery {
				return q.Since("applications", sinceTime)
			},
			expected: "createdDateTime ge " + sinceStr,
		},
		{
			name: "Until only",
			setup: func(q filterQuery) filterQuery {
				return q.Until("applications", untilTime)
			},
			expected: "createdDateTime le " + untilStr,
		},
		{
			name: "Select only",
			setup: func(q filterQuery) filterQuery {
				return q.Select("displayName eq 'Test'")
			},
			expected: "displayName eq 'Test'",
		},
		{
			name: "Since and Until",
			setup: func(q filterQuery) filterQuery {
				return q.Since("applications", sinceTime).Until("applications", untilTime)
			},
			expected: "createdDateTime ge " + sinceStr + " and createdDateTime le " + untilStr,
		},
		{
			name: "Since, Until and Select",
			setup: func(q filterQuery) filterQuery {
				return q.Since("applications", sinceTime).
					Until("applications", untilTime).
					Select("isDraft eq false")
			},
			expected: "createdDateTime ge " + sinceStr + " and createdDateTime le " + untilStr + " and isDraft eq false",
		},
		{
			name: "Multiple Selects",
			setup: func(q filterQuery) filterQuery {
				return q.Select("field1 eq 1").Select("field2 eq 2")
			},
			expected: "field1 eq 1 and field2 eq 2",
		},
		{
			name: "Zero time Since and Until",
			setup: func(q filterQuery) filterQuery {
				return q.Since("applications", time.Time{}).Until("applications", time.Time{})
			},
			expected: "",
		},
		{
			name: "Unknown object for Since and Until",
			setup: func(q filterQuery) filterQuery {
				return q.Since("unknown", sinceTime).Until("unknown", untilTime)
			},
			expected: "",
		},
		{
			name: "Empty filter",
			setup: func(q filterQuery) filterQuery {
				return q
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := tt.setup(filterQuery{})
			actual := res.String()

			compare := testutils.NewCompareResult()
			compare.Assert("filter query string", tt.expected, actual)
			compare.Validate(t, tt.name)
		})
	}
}
