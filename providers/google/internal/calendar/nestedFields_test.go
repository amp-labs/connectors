package calendar

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestJSONPathField(t *testing.T) {
	t.Parallel()

	assert.Equal(t, jsonPathField("start", "dateTime"), "$['start']['dateTime']")
	assert.Equal(t,
		jsonPathField("extendedProperties", "private", "com.acme.dealId"),
		"$['extendedProperties']['private']['com.acme.dealId']")
}

func TestNestRecordData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   map[string]any
		want    map[string]any
		wantErr error
	}{
		{
			name:  "plain keys are copied unchanged",
			input: map[string]any{"summary": "Standup", "start": map[string]any{"date": "2026-10-20"}},
			want:  map[string]any{"summary": "Standup", "start": map[string]any{"date": "2026-10-20"}},
		},
		{
			name: "JSONPath keys become nested objects",
			input: map[string]any{
				"summary":                "Standup",
				"$['start']['dateTime']": "2026-10-20T10:00:00Z",
				"$['start']['timeZone']": "UTC",
				"$['end']['dateTime']":   "2026-10-20T10:30:00Z",
			},
			want: map[string]any{
				"summary": "Standup",
				"start":   map[string]any{"dateTime": "2026-10-20T10:00:00Z", "timeZone": "UTC"},
				"end":     map[string]any{"dateTime": "2026-10-20T10:30:00Z"},
			},
		},
		{
			name: "JSONPath keys merge into an existing nested object",
			input: map[string]any{
				"start":                  map[string]any{"timeZone": "UTC"},
				"$['start']['dateTime']": "2026-10-20T10:00:00Z",
			},
			want: map[string]any{
				"start": map[string]any{"dateTime": "2026-10-20T10:00:00Z", "timeZone": "UTC"},
			},
		},
		{
			name: "keys containing dots stay a single segment",
			input: map[string]any{
				"$['extendedProperties']['private']['com.acme.dealId']": "42",
				"$['extendedProperties']['shared']['team']":             "sales",
			},
			want: map[string]any{
				"extendedProperties": map[string]any{
					"private": map[string]any{"com.acme.dealId": "42"},
					"shared":  map[string]any{"team": "sales"},
				},
			},
		},
		{
			name: "non-object value at intermediate path is a conflict",
			input: map[string]any{
				"start":                  "2026-10-20T10:00:00Z",
				"$['start']['dateTime']": "2026-10-20T10:00:00Z",
			},
			wantErr: errConflictingFields,
		},
		{
			name: "leaf set by both a nested object and a JSONPath key is a conflict",
			input: map[string]any{
				"start":                  map[string]any{"dateTime": "2026-10-20T10:00:00Z"},
				"$['start']['dateTime']": "2026-10-20T11:00:00Z",
			},
			wantErr: errConflictingFields,
		},
		{
			name: "array wildcards are rejected",
			input: map[string]any{
				"$['attendees'][*]['email']": "a@example.com",
			},
			wantErr: errFieldPathNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := nestRecordData(tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestSelectEventFields(t *testing.T) {
	t.Parallel()

	record := map[string]any{
		"id":      "evt-1",
		"summary": "Standup",
		"start":   map[string]any{"dateTime": "2026-10-20T10:00:00Z", "timeZone": "UTC"},
		"extendedProperties": map[string]any{
			"private": map[string]any{"com.acme.dealId": "42"},
		},
	}

	got := selectEventFields(record, []string{
		"summary",
		"$['start']['dateTime']",
		"$['start']['date']", // absent for timed events
		"$['extendedProperties']['private']['com.acme.dealId']",
	})

	assert.DeepEqual(t, got, map[string]any{
		"id":      "evt-1",
		"summary": "Standup",
		"start":   map[string]any{"datetime": "2026-10-20T10:00:00Z"},
		"extendedproperties": map[string]any{
			"private": map[string]any{"com.acme.dealid": "42"},
		},
	})
}
