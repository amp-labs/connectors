package readhelper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestSetNestedValueSparseParent fixes projection order explicitly so the null
// mutation regression cannot depend on Go's randomized field-set iteration.
func TestSetNestedValueSparseParent(t *testing.T) {
	t.Parallel()

	for _, parentFirst := range []bool{true, false} {
		name := "descendant first"
		if parentFirst {
			name = "parent first"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			record := map[string]any{"items": []any{nil, map[string]any{"label": "yes"}}}
			original := map[string]any{"items": []any{nil, map[string]any{"label": "yes"}}}
			output := make(map[string]any)
			parentPath := []string{"items"}
			childPath := []string{"items", ArrayWildIndex, "label"}

			if parentFirst {
				setNestedValue(output, parentPath, record["items"])
			}

			if value, exists := getNestedValue(record, childPath); exists {
				setNestedValue(output, childPath, value)
			}

			if !parentFirst {
				setNestedValue(output, parentPath, record["items"])
			}

			if diff := cmp.Diff(original, output); diff != "" {
				t.Errorf("selected parent changed (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(original, record); diff != "" {
				t.Errorf("source changed (-want +got):\n%s", diff)
			}
		})
	}
}
