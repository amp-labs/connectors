package codec // nolint:dupl,varnamelen

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDecoratedRecord(t *testing.T) {
	t.Parallel()

	type Address struct {
		PostalCode string `json:"postalCode,omitempty"`
		State      string `json:"state,omitempty"`
		Country    string `json:"country,omitempty"`
	}

	type User struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Address *Address `json:"address,omitempty"`
	}

	tests := []struct {
		name           string
		baseData       map[string]any
		decorationData any
		expectedJSON   map[string]any
	}{
		{
			name:           "Simple identity without base",
			baseData:       nil,
			decorationData: User{ID: "id_55", Name: "Alice"},
			expectedJSON: map[string]any{
				"id":   "id_55",
				"name": "Alice",
			},
		},
		{
			name:           "Fields from decoration are marshalled with the base",
			baseData:       map[string]any{"age": 18.0},
			decorationData: User{ID: "id_55", Name: "Alice"},
			expectedJSON: map[string]any{
				"id":   "id_55",
				"name": "Alice",
				"age":  18.0,
			},
		},
		{
			name:           "Decoration overrides fields in the base",
			baseData:       map[string]any{"age": 18.0, "name": "Alice"},
			decorationData: User{ID: "id_55", Name: "Bob"},
			expectedJSON: map[string]any{
				"id":   "id_55",
				"name": "Bob",
				"age":  18.0,
			},
		},
		{
			name:           "Decoration overrides fields in the base",
			baseData:       map[string]any{"age": 18.0, "name": "Alice"},
			decorationData: User{ID: "id_55", Name: "Bob"},
			expectedJSON: map[string]any{
				"id":   "id_55",
				"name": "Bob",
				"age":  18.0,
			},
		},
		{
			name: "Decoration with nested struct overrides fields in the appropriate level",
			baseData: map[string]any{
				"age":  18.0,
				"name": "Alice",
				"address": map[string]any{
					"postalCode": "ABC88", // must be in final output
					"state":      "California",
				},
			},
			decorationData: User{
				ID:   "id_55",
				Name: "Bob", // replaces
				Address: &Address{
					State:   "Colorado", // replaces
					Country: "USA",
				},
			},
			expectedJSON: map[string]any{
				"id":   "id_55",
				"name": "Bob",
				"age":  18.0,
				"address": map[string]any{
					"postalCode": "ABC88",
					"state":      "Colorado",
					"country":    "USA",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := NewDecoratedRecord(tt.baseData, tt.decorationData)

			result := testutils.NewCompareResult()
			defer result.Validate(t, tt.name)

			data, err := json.Marshal(obj)
			if !result.AssertErr("json.Marshal", nil, err) {
				return
			}

			registry := map[string]any{}
			err = json.Unmarshal(data, &registry)
			if !result.AssertErr("json.Unmarshal", nil, err) {
				return
			}

			result.Assert("Marshalled JSON", tt.expectedJSON, registry)
		})
	}
}
