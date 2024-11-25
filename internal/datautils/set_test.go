package datautils

import (
	"sort"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestStringSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty set",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Multiple elements",
			input:    []string{"apple", "pineapple"},
			expected: []string{"apple", "pineapple"},
		},
		{
			name:     "Item repetitions",
			input:    []string{"apple", "kiwi", "kiwi", "orange", "pineapple", "kiwi"},
			expected: []string{"apple", "kiwi", "orange", "pineapple"},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			set := NewStringSet(tt.input...)
			output := set.List()
			sort.Strings(output)
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, nil)
		})
	}
}
