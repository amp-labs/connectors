package utils

import (
	"errors"
	"reflect"
	"testing"
)

func TestSubstituteErrorsToStrings(t *testing.T) {
	errA := errors.New("A")
	errB := errors.New("B")

	type fruits struct {
		Name     string
		Warnings []error
	}

	type box struct {
		Content fruits
		Weight  int
	}

	tests := []struct {
		name string
		in   any
		want any
	}{
		{
			name: "Plain error",
			in:   errA,
			want: "A",
		},
		{
			name: "Struct with error field",
			in: struct {
				Message string
				Error   error
			}{"x", errA},
			want: map[string]any{
				"Message": "x",
				"Error":   "A",
			},
		},
		{
			name: "Slice of errors",
			in:   []error{errA, errB},
			want: []any{"A", "B"},
		},
		{
			name: "Map of errors",
			in: map[string]error{
				"a": errA,
				"b": errB,
			},
			want: map[string]any{
				"a": "A",
				"b": "B",
			},
		},
		{
			name: "Pointer to error",
			in:   &errA,
			want: "A",
		},
		{
			name: "Interface containing error",
			in:   any(errB),
			want: "B",
		},
		{
			name: "Nil input",
			in:   nil,
			want: nil,
		},
		{
			name: "Deeply nested struct with slice of errors",
			in: box{
				Content: fruits{
					Name:     "Apples",
					Warnings: []error{errors.New("too ripe"), errors.New("bruised")},
				},
				Weight: 5,
			},
			want: map[string]any{
				"Content": map[string]any{
					"Name":     "Apples",
					"Warnings": []any{"too ripe", "bruised"},
				},
				"Weight": 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteErrorsToStrings(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("substituteErrorsToStrings(%#v) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
