package datautils

import (
	"errors"
	"strconv"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestForEachWithErr(t *testing.T) { // nolint:funlen
	t.Parallel()

	type mapper func(string) (int, error)

	problem := errors.New("awesome custom error") // nolint:err113

	tests := []struct {
		name        string
		input       []string
		mapper      mapper
		expected    []int
		expectError error
	}{
		{
			name:  "Empty input slice",
			input: []string{},
			mapper: func(text string) (int, error) {
				return len(text), nil
			},
			expected:    []int{},
			expectError: nil,
		},
		{
			name:        "All valid conversions",
			input:       []string{"1", "2", "3"},
			mapper:      strconv.Atoi,
			expected:    []int{1, 2, 3},
			expectError: nil,
		},
		{
			name:  "Contains invalid conversion",
			input: []string{"10", "x", "30"},
			mapper: func(text string) (int, error) {
				num, err := strconv.Atoi(text)
				if err != nil {
					return 0, problem
				}

				return num, nil
			},
			expected:    nil,
			expectError: problem,
		},
		{
			name:  "Custom mapper that errors on specific input",
			input: []string{"a", "b", "error", "c"},
			mapper: func(text string) (int, error) {
				if text == "error" {
					return 0, problem
				}

				return len(text), nil
			},
			expected:    nil,
			expectError: problem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := ForEachWithErr(tt.input, tt.mapper)
			testutils.CheckOutputWithError(t, tt.name, tt.expected, tt.expectError, output, err)
		})
	}
}

func TestToAnySlice(t *testing.T) {
	t.Parallel()

	type order struct{ ID int }

	tests := []struct {
		name     string
		input    any
		expected []any
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []any{},
		},
		{
			name:     "Integers",
			input:    []int{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "Strings",
			input:    []string{"a", "b", "c"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "Booleans",
			input:    []bool{true, false, true},
			expected: []any{true, false, true},
		},
		{
			name:     "Structs",
			input:    []order{{1}, {2}},
			expected: []any{order{1}, order{2}},
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use generic that corresponds to each data type.
			var output []any
			switch v := tt.input.(type) {
			case []int:
				output = ToAnySlice(v)
			case []string:
				output = ToAnySlice(v)
			case []bool:
				output = ToAnySlice(v)
			case []order:
				output = ToAnySlice(v)
			default:
				t.Fatalf("unsupported test input type %T", v)
			}

			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, nil)
		})
	}
}
