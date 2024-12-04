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
		input    *common.ReadResult
		expected string
	}{
		{
			name:     "Nil Input",
			input:    nil,
			expected: "",
		},
		{
			name:     "Empty Data",
			input:    &common.ReadResult{Data: []common.ReadResultRow{}},
			expected: "",
		},
		{
			name: "Valid Fields[id]",
			input: &common.ReadResult{
				Data: []common.ReadResultRow{
					{Fields: map[string]any{string(ObjectFieldId): "12345"}},
				},
			},
			expected: "12345",
		},
		{
			name: "Valid Fields[hs_object_id]",
			input: &common.ReadResult{
				Data: []common.ReadResultRow{
					{Fields: map[string]any{string(ObjectFieldHsObjectId): "67890"}},
				},
			},
			expected: "67890",
		},
		{
			name: "Valid Raw[id]",
			input: &common.ReadResult{
				Data: []common.ReadResultRow{
					{Raw: map[string]any{string(ObjectFieldId): "abcdef"}},
				},
			},
			expected: "abcdef",
		},
		{
			name: "Valid Raw[properties][hs_object_id]",
			input: &common.ReadResult{
				Data: []common.ReadResultRow{
					{
						Raw: map[string]any{
							string(ObjectFieldProperties): map[string]any{
								string(ObjectFieldHsObjectId): "ghijkl",
							},
						},
					},
				},
			},
			expected: "ghijkl",
		},
		{
			name: "Multiple Records - Returns ID from the last row",
			input: &common.ReadResult{
				Data: []common.ReadResultRow{
					{Raw: map[string]any{string(ObjectFieldId): "first-id"}},
					{Raw: map[string]any{string(ObjectFieldProperties): map[string]any{string(ObjectFieldHsObjectId): "last-id"}}},
				},
			},
			expected: "last-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := GetLastResultId(test.input)
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}
