package monday

import (
	"testing"
)

func TestCustomFieldKey(t *testing.T) {
	t.Parallel()

	columnID := "status"
	if got := CustomFieldKey(columnID); got != "cf_status" {
		t.Fatalf("CustomFieldKey: got %q want cf_status", got)
	}
}

func TestItemReadFieldsIncludeColumnValues(t *testing.T) {
	t.Parallel()

	if !itemReadFieldsIncludeColumnValues([]string{"name", "cf_status"}) {
		t.Fatal("expected column_values inclusion for cf_ field")
	}

	if itemReadFieldsIncludeColumnValues([]string{"name", "id"}) {
		t.Fatal("did not expect column_values inclusion")
	}
}

func TestFlattenItemColumnValues(t *testing.T) {
	t.Parallel()

	obj := map[string]any{
		"id": "1",
		"column_values": []any{
			map[string]any{"id": "status", "text": "Done", "type": "status"},
			map[string]any{"id": "numbers", "text": "42", "type": "numbers"},
		},
	}

	flattenItemColumnValues(obj)

	if obj[CustomFieldKey("status")] != "Done" {
		t.Fatalf("status: got %#v", obj[CustomFieldKey("status")])
	}

	if obj[CustomFieldKey("numbers")] != "42" {
		t.Fatalf("numbers: got %#v", obj[CustomFieldKey("numbers")])
	}
}

func TestPrepareItemWriteRecordData(t *testing.T) {
	t.Parallel()

	columns := map[string]mondayColumnDefinition{
		"status": {ID: "status", Type: "status"},
		"text":   {ID: "text", Type: "text"},
	}

	record := map[string]any{
		"board_id": "123",
		"name":     "Task",
		"cf_status": "Working on it",
		"cf_text":   "hello",
	}

	prepared, err := prepareItemWriteRecordData(record, columns)
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
