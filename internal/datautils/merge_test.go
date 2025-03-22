package datautils

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestMergeUniqueLists(t *testing.T) { // nolint:funlen
	t.Parallel()

	type collection = UniqueLists[string, string]

	tests := []struct {
		name     string
		input    []collection
		expected collection
	}{
		{
			name: "Colliding names are merged",
			input: []collection{{
				"fruits": NewSet("apple", "banana"),
			}, {
				"fruits": NewSet("apple", "orange"), // apple already exists, orange is new
			}},
			expected: collection{
				"fruits": NewSet("apple", "banana", "orange"),
			},
		},
		{
			name: "Different named sets are propagated",
			input: []collection{{
				"numbers": NewSet("1", "2", "3"),
			}, {
				"letters": NewSet("A", "B", "C"), // new set
			}},
			expected: collection{
				"numbers": NewSet("1", "2", "3"),
				"letters": NewSet("A", "B", "C"),
			},
		},
		{
			name: "Combine diverse lists of sets",
			input: []collection{
				{
					"digits":  NewSet("1", "2", "3"),
					"decades": NewSet("10", "20"),
				},
				{
					"digits":  NewSet("3", "4"), // 3 is duplicate
					"decades": NewSet("30"),     // 30 is a new number
					"grades":  NewSet("A", "B"), // new set
				},
			},
			expected: collection{
				"digits":  NewSet("1", "2", "3", "4"),
				"decades": NewSet("10", "20", "30"),
				"grades":  NewSet("A", "B"),
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := MergeUniqueLists(tt.input...)

			testutils.CheckOutput(t, tt.name, tt.expected, output)
		})
	}
}
