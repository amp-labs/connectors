package phoneburner

import (
	"strings"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/spyzhov/ajson"
)

func TestParseMemberCustomFieldDefinitionsPage_skipsBlankDisplayName(t *testing.T) {
	t.Parallel()

	body, err := ajson.Unmarshal([]byte(`{
		"customfields": {
			"customfields": [
				{
					"custom_field_id": "1",
					"display_name": "",
					"type_id": "1",
					"type_name": "Text Field"
				},
				{
					"custom_field_id": "215",
					"display_name": "Lead Score",
					"type_id": "7",
					"type_name": "Numeric"
				}
			],
			"total_pages": 1
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}

	defs, totalPages, err := parseMemberCustomFieldDefinitionsPage(body)
	if err != nil {
		t.Fatal(err)
	}

	if totalPages != 1 {
		t.Fatalf("totalPages = %d, want 1", totalPages)
	}

	if len(defs) != 1 || defs[0].DisplayName != "Lead Score" {
		t.Fatalf("defs = %#v, want single Lead Score row", defs)
	}
}

func TestReadContactRecordTransformerAndMarshaledDataWithId(t *testing.T) {
	t.Parallel()

	leadKey := customFieldMetadataKey(leadScoreDisplayName)

	jsonStr := `{
		"contact_user_id": "30919237",
		"custom_fields": [
			{"name": "` + leadScoreDisplayName + `", "type": "7", "value": "42"}
		]
	}`

	node, err := ajson.Unmarshal([]byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	marshalFunc := readhelper.MakeMarshaledDataFuncWithId(
		readContactRecordTransformer,
		readhelper.NewIdField("contact_user_id"),
	)

	out, err := marshalFunc([]*ajson.Node{node}, connectors.Fields("contact_user_id", leadKey).List())
	if err != nil {
		t.Fatalf("MakeMarshaledDataFuncWithId: %v", err)
	}

	if len(out) != 1 {
		t.Fatalf("len(out) = %d", len(out))
	}

	row := out[0]

	if got, want := row.Id, "30919237"; got != want {
		t.Fatalf("Id = %q, want %q", got, want)
	}

	// Fields: flattened custom value (lookup key is lowercased by ExtractLowercaseFieldsFromRaw).
	fieldsKey := strings.ToLower(leadKey)
	if got, want := row.Fields[fieldsKey], "42"; got != want {
		t.Fatalf("Fields[%q] = %v, want %v", fieldsKey, got, want)
	}

	// Raw: must stay provider-shaped; no top-level merged custom key.
	if _, ok := row.Raw[leadKey]; ok {
		t.Fatalf(`Raw must not contain flattened connector key %q`, leadKey)
	}

	if _, ok := row.Raw[fieldsKey]; ok {
		t.Fatalf(`Raw must not contain Fields-style lowercased key %q`, fieldsKey)
	}

	cf, ok := row.Raw["custom_fields"]
	if !ok {
		t.Fatal(`Raw must still include provider "custom_fields" array`)
	}

	list, ok := cf.([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("Raw custom_fields = %v", cf)
	}
}
