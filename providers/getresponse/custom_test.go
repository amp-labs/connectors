package getresponse

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/brianvoe/gofakeit/v6"
)

func TestCustomFieldKey(t *testing.T) {
	t.Parallel()

	customFieldDefinitionID := "newsletter_subscription_tier"
	if got := CustomFieldKey(customFieldDefinitionID); got != "cf_"+customFieldDefinitionID {
		t.Fatalf("CustomFieldKey: got %q want cf_%s", got, customFieldDefinitionID)
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
		{
			name:  "cf_ prefixed field appends customFieldValues to query",
			input: []string{"email", "cf_newsletter_preference"},
			want:  []string{"email", "cf_newsletter_preference", "customFieldValues"},
		},
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

	contactEmail := gofakeit.Email()
	singleSelectFieldID := "loyalty_tier_level"
	multiSelectFieldID := "product_interest_tags"

	obj := map[string]any{
		"email": contactEmail,
		"customFieldValues": []any{
			map[string]any{"customFieldId": singleSelectFieldID, "value": []any{"gold_member"}},
			map[string]any{"customFieldId": multiSelectFieldID, "value": []any{"newsletter", "webinar"}},
		},
	}

	flattenContactCustomFieldValues(obj)

	if obj[CustomFieldKey(singleSelectFieldID)] != "gold_member" {
		t.Fatalf("%s: got %#v", singleSelectFieldID, obj[CustomFieldKey(singleSelectFieldID)])
	}

	multiValue, ok := obj[CustomFieldKey(multiSelectFieldID)].([]any)
	if !ok || len(multiValue) != 2 {
		t.Fatalf("%s: got %#v", multiSelectFieldID, obj[CustomFieldKey(multiSelectFieldID)])
	}
}

func TestPrepareContactWriteRecordData(t *testing.T) {
	t.Parallel()

	t.Run("record without cf_ keys is returned unchanged", func(t *testing.T) {
		t.Parallel()

		contactEmail := gofakeit.Email()
		in := map[string]any{"email": contactEmail}
		out := prepareContactWriteRecordData(in)
		if len(out) != 1 || out["email"] != contactEmail {
			t.Fatalf("got %#v", out)
		}
	})

	t.Run("cf_ prefixed keys are merged into customFieldValues array", func(t *testing.T) {
		t.Parallel()

		contactEmail := gofakeit.Email()
		jobTitleCustomFieldID := "job_title_custom_field"
		jobTitleValue := "Senior Marketing Manager"

		in := map[string]any{
			"email":                               contactEmail,
			CustomFieldKey(jobTitleCustomFieldID): jobTitleValue,
		}
		out := prepareContactWriteRecordData(in)
		if out["email"] != contactEmail {
			t.Fatalf("email stripped: %#v", out)
		}

		if _, ok := out[CustomFieldKey(jobTitleCustomFieldID)]; ok {
			t.Fatal("cf_ key should be removed from root record")
		}

		raw, ok := out["customFieldValues"].([]any)
		if !ok || len(raw) != 1 {
			t.Fatalf("customFieldValues: %#v", out["customFieldValues"])
		}

		entry, ok := raw[0].(map[string]any)
		if !ok || entry["customFieldId"] != jobTitleCustomFieldID {
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
