package connectwise

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestGetValueType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldType string
		entryType string
		expected  common.ValueType
	}{
		// Regression: ConnectWise "List" entry method is a single-select dropdown, not multi-select.
		// Values in read responses are always scalar (e.g. "false"), never arrays.
		{
			name:      "List Text is single select",
			fieldType: "Text",
			entryType: "List",
			expected:  common.ValueTypeSingleSelect,
		},
		{
			name:      "List Number is single select",
			fieldType: "Number",
			entryType: "List",
			expected:  common.ValueTypeSingleSelect,
		},
		{
			name:      "List Percent is single select",
			fieldType: "Percent",
			entryType: "List",
			expected:  common.ValueTypeSingleSelect,
		},
		{
			name:      "Option Text is single select",
			fieldType: "Text",
			entryType: "Option",
			expected:  common.ValueTypeSingleSelect,
		},
		{
			name:      "EntryField Text is string",
			fieldType: "Text",
			entryType: "EntryField",
			expected:  common.ValueTypeString,
		},
		{
			name:      "EntryField Number is float",
			fieldType: "Number",
			entryType: "EntryField",
			expected:  common.ValueTypeFloat,
		},
		{
			name:      "EntryField Percent is float",
			fieldType: "Percent",
			entryType: "EntryField",
			expected:  common.ValueTypeFloat,
		},
		{
			name:      "Checkbox is boolean",
			fieldType: "Checkbox",
			entryType: "EntryField",
			expected:  common.ValueTypeBoolean,
		},
		{
			name:      "Date is date",
			fieldType: "Date",
			entryType: "Date",
			expected:  common.ValueTypeDate,
		},
		{
			name:      "Hyperlink is string",
			fieldType: "Hyperlink",
			entryType: "EntryField",
			expected:  common.ValueTypeString,
		},
		{
			name:      "TextArea is string",
			fieldType: "TextArea",
			entryType: "EntryField",
			expected:  common.ValueTypeString,
		},
		{
			name:      "Button is other",
			fieldType: "Button",
			entryType: "Button",
			expected:  common.ValueTypeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field := modelCustomField{
				FieldTypeIdentifier: tt.fieldType,
				EntryTypeIdentifier: tt.entryType,
			}

			if got := field.getValueType(); got != tt.expected {
				t.Errorf("getValueType() with fieldType=%q entryType=%q: got %q, want %q",
					tt.fieldType, tt.entryType, got, tt.expected)
			}
		})
	}
}
