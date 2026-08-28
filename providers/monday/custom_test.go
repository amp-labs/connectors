package monday

import (
	"testing"

	"github.com/spyzhov/ajson"
)

func TestCustomFieldKey(t *testing.T) {
	t.Parallel()

	if got := CustomFieldKey("status"); got != "cf_status" {
		t.Fatalf("CustomFieldKey: got %q want cf_status", got)
	}
}

func TestItemReadCustomFieldsQueryNeedsColumnValues(t *testing.T) {
	t.Parallel()

	if !itemReadCustomFieldsQueryNeedsColumnValues([]string{"name", "cf_status"}) {
		t.Fatal("expected column_values inclusion for cf_ field")
	}

	if itemReadCustomFieldsQueryNeedsColumnValues([]string{"name", "id"}) {
		t.Fatal("did not expect column_values inclusion")
	}
}

func TestItemReadRecordCustomFieldsTransformer(t *testing.T) {
	t.Parallel()

	node, err := ajson.Unmarshal([]byte(`{
		"id": "1",
		"column_values": [
			{"id": "status", "text": "Done", "type": "status"},
			{"id": "numbers", "text": "42", "type": "numbers"}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}

	obj, err := itemReadRecordCustomFieldsTransformer(node)
	if err != nil {
		t.Fatal(err)
	}

	if obj["cf_status"] != "Done" {
		t.Fatalf("status: got %#v", obj["cf_status"])
	}

	if obj["cf_numbers"] != "42" {
		t.Fatalf("numbers: got %#v", obj["cf_numbers"])
	}
}

func TestPrepareItemWriteCustomFieldsRecordData(t *testing.T) {
	t.Parallel()

	columns := map[string]mondayColumnDefinition{
		"status": {ID: "status", Type: "status"},
		"text":   {ID: "text", Type: "text"},
	}

	record := map[string]any{
		"board_id":  "123",
		"name":      "Task",
		"cf_status": "Working on it",
		"cf_text":   "hello",
	}

	prepared, err := prepareItemWriteCustomFieldsRecordData(record, columns)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := prepared["cf_status"]; ok {
		t.Fatal("cf_ keys should be removed from payload")
	}

	raw, ok := prepared["column_values"].(string)
	if !ok || raw == "" {
		t.Fatalf("column_values: got %#v", prepared["column_values"])
	}

	if prepared["name"] != "Task" {
		t.Fatalf("name: got %#v", prepared["name"])
	}
}

func TestBoardIDFromFilterString(t *testing.T) {
	t.Parallel()

	if got := boardIDFromFilterString("board_id=999&other=x"); got != "999" {
		t.Fatalf("got %q", got)
	}
}

func TestParseObjectNameAndBoardID(t *testing.T) {
	t.Parallel()

	obj, board := parseObjectNameAndBoardID("items@555")
	if obj != mondayObjectItems || board != "555" {
		t.Fatalf("got %q %q", obj, board)
	}
}
