package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

// nolint:funlen
func TestGetLastResultId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *common.ReadResultRow
		expected string
	}{
		{
			name:     "Nil Input",
			input:    nil,
			expected: "",
		},
		{
			name:     "Empty Data",
			input:    &common.ReadResultRow{},
			expected: "",
		},
		{
			name: "Valid Fields[id]",
			input: &common.ReadResultRow{
				Fields: map[string]any{string(ObjectFieldId): "12345"},
			},
			expected: "12345",
		},
		{
			name: "Valid Fields[hs_object_id]",
			input: &common.ReadResultRow{
				Fields: map[string]any{string(ObjectFieldHsObjectId): "67890"},
			},
			expected: "67890",
		},
		{
			name: "Valid Raw[id]",
			input: &common.ReadResultRow{
				Raw: map[string]any{string(ObjectFieldId): "abcdef"},
			},
			expected: "abcdef",
		},
		{
			name: "Valid Raw[properties][hs_object_id]",
			input: &common.ReadResultRow{
				Raw: map[string]any{
					string(ObjectFieldProperties): map[string]any{
						string(ObjectFieldHsObjectId): "ghijkl",
					},
				},
			},
			expected: "ghijkl",
		},
		{
			name: "Dummy hubspot test",
			input: &common.ReadResultRow{
				Fields: map[string]any{
					"lifecyclestage": "lead",
				},
				Raw: map[string]any{
					"archived":  false,
					"createdAt": "2010-12-08T06:13:17.698Z",
					"id":        "15237",
					"properties": map[string]any{
						"createdate":       "2010-12-08T06:13:17.698Z",
						"hs_object_id":     "15237",
						"lastmodifieddate": "2010-12-04T11:18:28.697Z",
						"lifecyclestage":   "lead",
					},
					"updatedAt": "2010-12-04T11:18:28.697Z",
				},
			},
			expected: "15237",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := GetResultId(test.input)
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}
