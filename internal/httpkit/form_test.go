package httpkit

import (
	"net/url"
	"reflect"
	"testing"
)

func TestEncodeForm(t *testing.T) { //nolint:funlen
	t.Parallel()

	encoded, err := EncodeForm(map[string]any{
		"z":   "last",
		"a":   "first",
		"nil": nil,
		"n":   10,
		"ss":  []string{"x", "y"},
		"arr": []any{"v1", map[string]any{"k": "v"}},
		"obj": map[string]any{"hello": "world"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := url.ParseQuery(string(encoded))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	expected := url.Values{
		"a":   []string{"first"},
		"arr": []string{"v1", `{"k":"v"}`},
		"n":   []string{"10"},
		"obj": []string{`{"hello":"world"}`},
		"ss":  []string{"x", "y"},
		"z":   []string{"last"},
	}

	if !reflect.DeepEqual(parsed, expected) {
		t.Fatalf("unexpected parsed values:\nexpected: %#v\ngot:      %#v", expected, parsed)
	}
}

