package getresponse

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestCustomFieldKey(t *testing.T) {
	t.Parallel()

	if got := CustomFieldKey("abc"); got != "cf_abc" {
		t.Fatalf("CustomFieldKey: got %q want cf_abc", got)
	}
}

func TestContactReadFieldsQueryForAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "empty", input: nil, want: nil},
		{name: "no custom", input: []string{"email", "name"}, want: []string{"email", "name"}},
		{name: "customFieldValues passthrough", input: []string{"email", "customFieldValues"}, want: []string{"email", "customFieldValues"}},
		{name: "folded customFieldValues", input: []string{"CustomFieldValues"}, want: []string{"CustomFieldValues"}},
		{name: "cf prefix appends", input: []string{"email", "cf_x1"}, want: []string{"email", "cf_x1", "customFieldValues"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := contactReadFieldsQueryForAPI(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("len got %d want %d: %#v vs %#v", len(got), len(tt.want), got, tt.want)
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("idx %d: got %q want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFlattenContactCustomFieldValues(t *testing.T) {
	t.Parallel()

	obj := map[string]any{
		"email": "a@b.c",
		"customFieldValues": []any{
			map[string]any{"customFieldId": "f1", "value": []any{"solo"}},
			map[string]any{"customFieldId": "f2", "value": []any{"a", "b"}},
		},
	}

	flattenContactCustomFieldValues(obj)

	if obj[CustomFieldKey("f1")] != "solo" {
		t.Fatalf("f1: got %#v", obj[CustomFieldKey("f1")])
	}

	v2, ok := obj[CustomFieldKey("f2")].([]any)
	if !ok || len(v2) != 2 {
		t.Fatalf("f2: got %#v", obj[CustomFieldKey("f2")])
	}
}

func TestMergeContactCustomFieldValuesIntoBody(t *testing.T) {
	t.Parallel()

	t.Run("no cf keys unchanged", func(t *testing.T) {
		t.Parallel()

		in := map[string]any{"email": "x@y.z"}
		out := mergeContactCustomFieldValuesIntoBody(in)
		if len(out) != 1 || out["email"] != "x@y.z" {
			t.Fatalf("got %#v", out)
		}
	})

	t.Run("cf keys become array", func(t *testing.T) {
		t.Parallel()

		in := map[string]any{
			"email":              "x@y.z",
			CustomFieldKey("a"): "hello",
		}
		out := mergeContactCustomFieldValuesIntoBody(in)
		if out["email"] != "x@y.z" {
			t.Fatalf("email stripped: %#v", out)
		}

		if _, ok := out[CustomFieldKey("a")]; ok {
			t.Fatal("cf key should be removed from root")
		}

		raw, ok := out["customFieldValues"].([]any)
		if !ok || len(raw) != 1 {
			t.Fatalf("customFieldValues: %#v", out["customFieldValues"])
		}

		m, ok := raw[0].(map[string]any)
		if !ok || m["customFieldId"] != "a" {
			t.Fatalf("entry: %#v", raw[0])
		}
	})
}

func TestGetResponseCustomField_fieldMetadata(t *testing.T) {
	t.Parallel()

	d := getResponseCustomField{
		CustomFieldId: "id1",
		Name:          "Job title",
		ValueType:     "string",
		Values:        []string{"x", "y"},
	}

	md := d.fieldMetadata()
	if md.DisplayName != "Job title" {
		t.Fatalf("DisplayName: %q", md.DisplayName)
	}

	if md.ValueType != common.ValueTypeString {
		t.Fatalf("ValueType: %v", md.ValueType)
	}

	if md.IsCustom == nil || !*md.IsCustom {
		t.Fatal("IsCustom should be true")
	}

	if len(md.Values) != 2 {
		t.Fatalf("Values len: %d", len(md.Values))
	}
}
