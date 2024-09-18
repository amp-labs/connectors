package handy

import (
	"testing"
	"time"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestTimingFormatRFC3339inUTC(t *testing.T) {
	t.Parallel()

	const hour = 60 * 60

	zoneUTC := time.FixedZone("UTC", 0)
	zonePacific := time.FixedZone("UTC-8", -8*hour)
	zoneEasternEu := time.FixedZone("UTC+2", 2*hour)

	createTimeIn := func(zone *time.Location) time.Time {
		return time.Date(2024, 9, 19, 4, 30, 45, 600, zone)
	}

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Time origin",
			input:    time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
		{
			name:     "UTC time",
			input:    createTimeIn(zoneUTC),
			expected: "2024-09-19T04:30:45Z",
		},
		{
			name:     "Pacific time",
			input:    createTimeIn(zonePacific),
			expected: "2024-09-19T12:30:45Z",
		},
		{
			name:     "Eastern EU time",
			input:    createTimeIn(zoneEasternEu),
			expected: "2024-09-19T02:30:45Z",
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := Time.FormatRFC3339inUTC(tt.input)
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, nil)
		})
	}
}
