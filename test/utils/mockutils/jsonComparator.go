package mockutils

import (
	"encoding/json"
	"reflect"
)

// JSONComparator offers helpers for structural equality checks between
// arbitrary Go values and JSON documents.
//
// It provides a robust comparison mechanism for verifying JSON
// equivalence in tests, ignoring superficial differences such as
// whitespace, field order, or formatting.
//
// Example:
//
//	type response struct {
//		ID   string `json:"id"`
//		Name string `json:"name"`
//	}
//
//	actual := response{ID: "123", Name: "Alice"}
//	expected := mockutils.JSONErrorWrapper(`{"name":"Alice","id":"123"}`)
//
//	ok := mockutils.JSONComparator.Equals(actual, expected)
//	// ok == true (field order does not matter)
//
// This utility is primarily used by ErrorNormalizedComparator
// to compare JSON-encoded error objects but can be reused
// directly in tests where normalized JSON equality is needed.
var JSONComparator = jsonComparator{}

type jsonComparator struct{}

// Equals reports whether the given Go value and a JSON-encoded
// expected representation are structurally equivalent.
//
// It returns true if both JSON documents represent identical key-value
// structures, ignoring field order and insignificant formatting.
//
// Any marshal or unmarshal failure results in false.
func (jsonComparator) Equals(actual any, expected string) bool {
	aJSON, err := json.Marshal(actual)
	if err != nil {
		return false
	}

	if jsonBodyMatch(aJSON, expected) {
		return true
	}

	return false
}

func jsonBodyMatch(actual []byte, expected string) bool {
	first := make(map[string]any)
	if err := json.Unmarshal(actual, &first); err != nil {
		return false
	}

	second := make(map[string]any)
	if err := json.Unmarshal([]byte(expected), &second); err != nil {
		return false
	}

	return reflect.DeepEqual(first, second)
}
