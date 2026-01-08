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
			name: "Object with custom fields should flatten custom fields to root level",
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
				"id":    "cus_test123",
				"email": "test@example.com",
				"name":  "Test Customer",
				"metadata": map[string]any{
					"order_id":     "6735",
					"user_id":      "456",
					"internal_ref": "REF-2024-001",
				},
				"order_id":     "6735",
				"user_id":      "456",
				"internal_ref": "REF-2024-001",
			},
			wantErr: false,
		},
		{
			name: "Object without custom fields should return as is",
			input: map[string]any{
				"id":    "cus_test123",
				"email": "test@example.com",
				"name":  "Test Customer",
			},
			expected: map[string]any{
				"id":    "cus_test123",
				"email": "test@example.com",
				"name":  "Test Customer",
			},
			wantErr: false,
		},
		{
			name: "Object with empty custom fields should return as is",
			input: map[string]any{
				"id":       "cus_test123",
				"email":    "test@example.com",
				"metadata": map[string]any{},
			},
			expected: map[string]any{
				"id":       "cus_test123",
				"email":    "test@example.com",
				"metadata": map[string]any{},
			},
			wantErr: false,
		},
		{
			name: "Object with non-map custom fields should return as is",
			input: map[string]any{
				"id":       "cus_test123",
				"email":    "test@example.com",
				"metadata": "not-a-map",
			},
			expected: map[string]any{
				"id":       "cus_test123",
				"email":    "test@example.com",
				"metadata": "not-a-map",
			},
			wantErr: false,
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

			result, err := flattenCustomFields(node)

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
