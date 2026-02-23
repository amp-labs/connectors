package kit

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
				"id":            "3837707452",
				"email_address": "test@example.com",
				"name":          "Test Customer",
				"fields": map[string]any{
					"animal": "dog",
				},
			},
			expected: map[string]any{
				"id":            "3837707452",
				"email_address": "test@example.com",
				"name":          "Test Customer",
				"fields": map[string]any{
					"animal": "dog",
				},
				"animal": "dog",
			},
			wantErr: false,
		},
		{
			name: "Object without custom fields should return as is",
			input: map[string]any{
				"id":            "3837707452",
				"email_address": "test@example.com",
				"name":          "Test Customer",
			},
			expected: map[string]any{
				"id":            "3837707452",
				"email_address": "test@example.com",
				"name":          "Test Customer",
			},
			wantErr: false,
		},
		{
			name: "Object with empty custom fields should return as is",
			input: map[string]any{
				"id":            "3837707452",
				"email_address": "test@example.com",
				"fields":        map[string]any{},
			},
			expected: map[string]any{
				"id":            "3837707452",
				"email_address": "test@example.com",
				"fields":        map[string]any{},
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
