package mockutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/test/utils/testutils"
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
func (errorNormalizedComparator) ErrorEquals(actualErr, expectedErr any) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	// 1. Direct equality first.
	if reflect.DeepEqual(actualErr, expectedErr) {
		return result // good
	}

	// 2. If both implement error, compare semantically.
	aErr, aOK := actualErr.(error)
	eErr, eOL := expectedErr.(error)

	if aOK && eOL {
		if errors.Is(aErr, eErr) || strings.Contains(aErr.Error(), eErr.Error()) {
			return result // good
		}

		result.Assert("error", eErr, aErr)
		return result
	}

	// 3. Handle JSON case if expected is a JSON string.
	if expectedJSON, ok := expectedErr.(JSONErrorWrapper); ok {
		aJSON, err := json.Marshal(actualErr)
		if err != nil {
			result.AddDiff("failed to marshal actual error to JSON: %v", err)
			return result
		}

		if jsonBodyMatch(aJSON, string(expectedJSON)) {
			return result // good
		}

		result.Assert("JSON error", string(expectedJSON), string(aJSON))
		return result
	}

	// 4. Fallback string-based comparison.
	aStr := fmt.Sprintf("%v", actualErr)
	eStr := fmt.Sprintf("%v", expectedErr)

	result.Assert("error string", eStr, aStr)

	return result
}

// EachErrorEquals compares two slices of heterogeneous error values.
// It returns true if each corresponding pair of elements is considered equal
// according to ErrorEquals.
//
// Order and slice length must match exactly.
func (c errorNormalizedComparator) EachErrorEquals(actual, expected []any) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	if len(actual) != len(expected) {
		result.AddDiff("expected %d errors, got %d", len(expected), len(actual))
		return result
	}

	for i := range len(actual) {
		res := c.ErrorEquals(actual[i], expected[i])

		for _, diff := range res.Diff {
			result.AddDiff("Errors[%d] %s", i, diff)
		}
	}

	return result
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
