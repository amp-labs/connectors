package mockutils

import (
	"encoding/json"

	"github.com/amp-labs/connectors/test/utils/testutils"
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
func (jsonComparator) Equals(expected string, actual any) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	actualBytes, err := json.Marshal(actual)
	if err != nil {
		result.AddDiff("jsonComparator.Equals: couldn't marshall: %v", actual)
		return result
	}

	actualJSON := make(map[string]any)
	if err = json.Unmarshal(actualBytes, &actualJSON); err != nil {
		result.AddDiff("jsonComparator.Equals: couldn't unmarshall 'actual' into map[string]any")
		return result
	}

	expectedJSON := make(map[string]any)
	if err = json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		result.AddDiff("jsonComparator.Equals: couldn't unmarshall 'expected' into map[string]any")
		return result
	}

	result.Assert("jsonComparator.Equals", actualJSON, expectedJSON)

	return result
}

func (jsonComparator) ListsEqual(expected []string, actual []any) (result *testutils.CompareResult) {
	result = testutils.NewCompareResult()

	if !result.Assert("jsonComparator.ListsEqual lists are of different sizes", len(expected), len(actual)) {
		return result
	}

	for index, e := range expected {
		a := actual[index]
		result.Merge(JSONComparator.Equals(e, a))
	}

	return result
}
