package readhelper

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/google/go-cmp/cmp"
)

func TestSelectFieldsSparseParent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record string
		field  string
	}{
		{
			name:   "null array item stays null",
			record: `{"items":[null,{"label":"yes"}]}`,
			field:  "$['items'][*]['label']",
		},
		{
			name:   "mixed case matched item keeps its original keys",
			record: `{"items":[{},{"Name":"yes"}]}`,
			field:  "$['items'][*]['Name']",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var record, original map[string]any
			if err := json.Unmarshal([]byte(tt.record), &record); err != nil {
				t.Fatal(err)
			}

			if err := json.Unmarshal([]byte(tt.record), &original); err != nil {
				t.Fatal(err)
			}

			got := SelectFields(record, datautils.NewStringSet("$['items']", tt.field))
			if diff := cmp.Diff(original, got); diff != "" {
				t.Errorf("selected parent changed (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(original, record); diff != "" {
				t.Errorf("source changed (-want +got):\n%s", diff)
			}
		})
	}
}

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
