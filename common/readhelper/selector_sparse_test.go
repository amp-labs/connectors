package readhelper

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/google/go-cmp/cmp"
)

func TestSelectFieldsSparseArrays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		record   string
		fields   []string
		expected string
	}{
		{
			name:     "unsupported terminal wildcard does not suppress its descendant",
			record:   `{"items":[{"label":"yes"}]}`,
			fields:   []string{"$['items'][*]", "$['items'][*]['label']"},
			expected: `{"items":[{"label":"yes"}]}`,
		},
		{
			name:     "unsupported nested terminal wildcard does not suppress its descendant",
			record:   `{"items":[{"tags":[{"label":"yes"}]}]}`,
			fields:   []string{"$['items'][*]['tags'][*]", "$['items'][*]['tags'][*]['label']"},
			expected: `{"items":[{"tags":[{"label":"yes"}]}]}`,
		},
		{
			name:     "missing first item keeps later matches at their original positions",
			record:   `{"items":[{"id":"a"},{"id":"b","label":"second"},{"id":"c","label":null}]}`,
			fields:   []string{"$['items'][*]['id']", "$['items'][*]['label']"},
			expected: `{"items":[{"id":"a"},{"id":"b","label":"second"},{"id":"c","label":null}]}`,
		},
		{
			name:     "missing middle and last items do not shift matched fields",
			record:   `{"items":[{"label":"first"},{"other":true},{"label":"third"},{}]}`,
			fields:   []string{"$['items'][*]['label']"},
			expected: `{"items":[{"label":"first"},{},{"label":"third"},{}]}`,
		},
		{
			name:     "nested sparse arrays preserve sibling alignment",
			record:   `{"items":[{"id":"a","tags":[{"value":"red"},{}]},{"id":"b"},{"id":"c","tags":[{}, {"value":null}]}]}`,
			fields:   []string{"$['items'][*]['id']", "$['items'][*]['tags'][*]['value']"},
			expected: `{"items":[{"id":"a","tags":[{"value":"red"},{}]},{"id":"b"},{"id":"c","tags":[{},{"value":null}]}]}`,
		},
		{
			name:     "empty arrays stay empty alongside missing paths",
			record:   `{"items":[{}, {"tags":[]}]}`,
			fields:   []string{"$['items'][*]['tags'][*]['value']"},
			expected: `{"items":[{}, {"tags":[]}]}`,
		},
		{
			name:     "unselected null item keeps an empty projection slot",
			record:   `{"items":[null,{"label":"yes"}]}`,
			fields:   []string{"$['items'][*]['label']"},
			expected: `{"items":[{},{"label":"yes"}]}`,
		},
		{
			name:     "no matching item omits the selected field",
			record:   `{"items":[{}, {"other":true}]}`,
			fields:   []string{"$['items'][*]['label']"},
			expected: `{}`,
		},
		{
			name:     "wrong type in one item does not discard other matches",
			record:   `{"items":[{"details":{"label":"first"}},{"details":null},{"details":"text"}]}`,
			fields:   []string{"$['items'][*]['details']['label']"},
			expected: `{"items":[{"details":{"label":"first"}},{},{}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var record, expected, original map[string]any
			for _, input := range []struct {
				text string
				into *map[string]any
			}{{tt.record, &record}, {tt.expected, &expected}, {tt.record, &original}} {
				if err := json.Unmarshal([]byte(input.text), input.into); err != nil {
					t.Fatal(err)
				}
			}

			got := SelectFields(record, datautils.NewSetFromList(tt.fields))
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("projection mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(original, record); diff != "" {
				t.Errorf("source changed (-want +got):\n%s", diff)
			}
		})
	}
}

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
