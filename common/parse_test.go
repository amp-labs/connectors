package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMarshalledData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		records    []map[string]any
		fields     []string
		expectedId string
	}{
		{
			name: "String id is extracted as-is",
			records: []map[string]any{
				{"id": "abc-123", "name": "Alice"},
			},
			fields:     []string{"name"},
			expectedId: "abc-123",
		},
		{
			name: "Float64 id is converted to string",
			records: []map[string]any{
				{"id": float64(12345), "name": "Bob"},
			},
			fields:     []string{"name"},
			expectedId: "12345",
		},
		{
			name: "Float64 decimal id preserves precision",
			records: []map[string]any{
				{"id": float64(99.5), "name": "Carol"},
			},
			fields:     []string{"name"},
			expectedId: "99.5",
		},
		{
			name: "json.Number id is converted to string",
			records: []map[string]any{
				{"id": json.Number("67890"), "name": "Dave"},
			},
			fields:     []string{"name"},
			expectedId: "67890",
		},
		{
			name: "Missing id results in empty string",
			records: []map[string]any{
				{"name": "Eve"},
			},
			fields:     []string{"name"},
			expectedId: "",
		},
		{
			name: "Unsupported id type results in empty string",
			records: []map[string]any{
				{"id": true, "name": "Frank"},
			},
			fields:     []string{"name"},
			expectedId: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := GetMarshalledData(tt.records, tt.fields)
			require.NoError(t, err)
			require.Len(t, result, 1)
			assert.Equal(t, tt.expectedId, result[0].Id)
		})
	}
}

func TestGetMarshalledDataMultipleRecords(t *testing.T) {
	t.Parallel()

	records := []map[string]any{
		{"id": "str-1", "name": "Alice"},
		{"id": float64(200), "name": "Bob"},
		{"id": json.Number("300"), "name": "Carol"},
	}

	result, err := GetMarshalledData(records, []string{"name"})
	require.NoError(t, err)
	require.Len(t, result, 3)

	assert.Equal(t, "str-1", result[0].Id)
	assert.Equal(t, "200", result[1].Id)
	assert.Equal(t, "300", result[2].Id)
}
