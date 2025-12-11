package mockutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrorNormalizedComparator provides helper methods to compare error values
// of arbitrary types (errors, structs, strings, or JSON).
//
// It performs flexible equality checks that normalize differences in format,
// allowing robust test assertions across heterogeneous error types.
var ErrorNormalizedComparator = errorNormalizedComparator{}

type errorNormalizedComparator struct{}

// ErrorEquals compares two arbitrary error representations for semantic equality.
//
// Comparison is performed in the following order:
//  1. Direct equality via reflect.DeepEqual.
//  2. If both values implement error, compare using errors.Is or substring match.
//  3. If the expected value is a JSONErrorWrapper, compare by marshaling the
//     actual value to JSON and checking structural equality.
//  4. Fallback: string comparison via fmt.Sprintf("%v").
//
// It returns true if the two values are considered equivalent under these rules.
func (errorNormalizedComparator) ErrorEquals(actualErr, expectedErr any) bool {
	// 1. Direct equality first.
	if reflect.DeepEqual(actualErr, expectedErr) {
		return true
	}

	// 2. If both implement error, compare semantically.
	aErr, aOK := actualErr.(error)
	eErr, eOL := expectedErr.(error)

	if aOK && eOL {
		if errors.Is(aErr, eErr) || strings.Contains(aErr.Error(), eErr.Error()) {
			return true
		}

		return false
	}

	// 3. Handle JSON case if expected is a JSON string.
	if expectedJSON, ok := expectedErr.(JSONErrorWrapper); ok {
		aJSON, err := json.Marshal(actualErr)
		if err != nil {
			return false
		}

		if jsonBodyMatch(aJSON, string(expectedJSON)) {
			return true
		}

		return false
	}

	// 4. Fallback string-based comparison.
	aStr := fmt.Sprintf("%v", actualErr)
	eStr := fmt.Sprintf("%v", expectedErr)

	if aStr == eStr {
		return true
	}

	return false
}

// EachErrorEquals compares two slices of heterogeneous error values.
// It returns true if each corresponding pair of elements is considered equal
// according to ErrorEquals.
//
// Order and slice length must match exactly.
func (c errorNormalizedComparator) EachErrorEquals(actual, expected []any) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i := range len(actual) {
		if !c.ErrorEquals(actual[i], expected[i]) {
			return false
		}
	}

	return true
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
