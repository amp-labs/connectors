package phoneburner

import (
	"testing"

	"github.com/amp-labs/connectors"
)

func TestGetMarshaledDataContactsWithCustomFieldsPreservingRaw(t *testing.T) {
	t.Parallel()

	const leadKey = "custom_Lead Score" // metadata / read request field key (prefix + display name)

	record := map[string]any{
		"contact_user_id": "30919237",
		"custom_fields": []any{
			map[string]any{
				"name":  "Lead Score",
				"type":  "7",
				"value": "42",
			},
		},
	}

	records := []map[string]any{record}

	out, err := getMarshaledDataContactsWithCustomFieldsPreservingRaw(
		records,
		connectors.Fields("contact_user_id", leadKey).List(),
	)
	if err != nil {
		t.Fatalf("getMarshaledDataContactsWithCustomFieldsPreservingRaw: %v", err)
	}

	if len(out) != 1 {
		t.Fatalf("len(out) = %d", len(out))
	}

	row := out[0]

	// Fields: flattened custom value (lookup key is lowercased by ExtractLowercaseFieldsFromRaw).
	if got, want := row.Fields["custom_lead score"], "42"; got != want {
		t.Fatalf("Fields[custom_lead score] = %v, want %v", got, want)
	}

	// Raw: must stay provider-shaped; no top-level merged custom key (that would mean we confused
	// Fields with Raw, e.g. raw["custom_lead_score"] or raw["custom_Lead Score"]).
	if _, ok := row.Raw["custom_Lead Score"]; ok {
		t.Fatal(`Raw must not contain flattened custom field key "custom_Lead Score"`)
	}

	if _, ok := row.Raw["custom_lead score"]; ok {
		t.Fatal(`Raw must not contain Fields-style lowercased key "custom_lead score"`)
	}

	cf, ok := row.Raw["custom_fields"]
	if !ok {
		t.Fatal(`Raw must still include provider "custom_fields" array`)
	}

	list, ok := cf.([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("Raw custom_fields = %v", cf)
	}

	// Input row from extract must be unchanged (we only flatten the working copy).
	if _, still := record["custom_fields"]; !still {
		t.Fatal("extracted map must still have custom_fields after marshal")
	}

	if _, bad := record["custom_Lead Score"]; bad {
		t.Fatal("extracted map must not be flattened in place")
	}
}
