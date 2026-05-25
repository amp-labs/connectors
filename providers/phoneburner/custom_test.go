package phoneburner

import (
	"testing"

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
