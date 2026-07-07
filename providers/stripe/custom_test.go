package stripe

import (
	"encoding/json"
	"testing"

	"github.com/spyzhov/ajson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenCustomFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
		wantErr  bool
	}{
		{
			name: "Object with custom fields should return them",
			input: map[string]any{
				"id":    "cus_test123",
				"email": "test@example.com",
				"name":  "Test Customer",
				"metadata": map[string]any{
					"order_id":     "6735",
					"user_id":      "456",
					"internal_ref": "REF-2024-001",
				},
			},
			expected: map[string]any{
				"order_id":     "6735",
				"user_id":      "456",
				"internal_ref": "REF-2024-001",
			},
			wantErr: false,
		},
		{
			name: "Object without custom fields should return nothing",
			input: map[string]any{
				"id":    "cus_test123",
				"email": "test@example.com",
				"name":  "Test Customer",
			},
			expected: make(map[string]any),
			wantErr:  false,
		},
		{
			name: "Object with empty custom fields should return nothing",
			input: map[string]any{
				"id":       "cus_test123",
				"email":    "test@example.com",
				"metadata": map[string]any{},
			},
			expected: make(map[string]any),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jsonBytes, err := json.Marshal(tt.input)
			require.NoError(t, err)

			node, err := ajson.Unmarshal(jsonBytes)
			require.NoError(t, err)

			result, err := getCustomFields(node)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
