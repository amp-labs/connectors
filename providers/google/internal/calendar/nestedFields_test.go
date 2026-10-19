package calendar

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestNestRecordData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{
			name:  "plain keys are copied unchanged",
			input: map[string]any{"summary": "Standup", "start": map[string]any{"date": "2026-10-20"}},
			want:  map[string]any{"summary": "Standup", "start": map[string]any{"date": "2026-10-20"}},
		},
		{
			name: "dotted keys become nested objects",
			input: map[string]any{
				"summary":        "Standup",
				"start.dateTime": "2026-10-20T10:00:00Z",
				"start.timeZone": "UTC",
				"end.dateTime":   "2026-10-20T10:30:00Z",
			},
			want: map[string]any{
				"summary": "Standup",
				"start":   map[string]any{"dateTime": "2026-10-20T10:00:00Z", "timeZone": "UTC"},
				"end":     map[string]any{"dateTime": "2026-10-20T10:30:00Z"},
			},
		},
		{
			name: "dotted keys merge into an existing nested object",
			input: map[string]any{
				"start":          map[string]any{"timeZone": "UTC"},
				"start.dateTime": "2026-10-20T10:00:00Z",
			},
			want: map[string]any{
				"start": map[string]any{"dateTime": "2026-10-20T10:00:00Z", "timeZone": "UTC"},
			},
		},
		{
			name: "deep paths such as extended properties",
			input: map[string]any{
				"extendedProperties.private.crmId": "42",
				"extendedProperties.shared.team":   "sales",
			},
			want: map[string]any{
				"extendedProperties": map[string]any{
					"private": map[string]any{"crmId": "42"},
					"shared":  map[string]any{"team": "sales"},
				},
			},
		},
		{
			name: "non-object value at intermediate path is replaced",
			input: map[string]any{
				"start":          "2026-10-20T10:00:00Z",
				"start.dateTime": "2026-10-20T10:00:00Z",
			},
			want: map[string]any{
				"start": map[string]any{"dateTime": "2026-10-20T10:00:00Z"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.DeepEqual(t, nestRecordData(tt.input), tt.want)
		})
	}
}

func TestMarshalRecordsWithNestedFields(t *testing.T) {
	t.Parallel()

	record := map[string]any{
		"id":      "evt-1",
		"summary": "Standup",
		"start":   map[string]any{"dateTime": "2026-10-20T10:00:00Z", "timeZone": "UTC"},
		"extendedProperties": map[string]any{
			"private": map[string]any{"crmId": "42"},
		},
	}

	rows, err := marshalRecordsWithNestedFields([]map[string]any{record}, []string{
		"summary",
		"start.dateTime",
		"START.TIMEZONE",
		"start.date", // absent for timed events
		"extendedProperties.private.crmId",
	})
	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)

	assert.DeepEqual(t, rows[0].Fields, map[string]any{
		"id":                               "evt-1",
		"summary":                          "Standup",
		"start.datetime":                   "2026-10-20T10:00:00Z",
		"start.timezone":                   "UTC",
		"extendedproperties.private.crmid": "42",
	})
	// Raw keeps the provider shape without synthetic dotted keys.
	assert.DeepEqual(t, rows[0].Raw, record)
}
