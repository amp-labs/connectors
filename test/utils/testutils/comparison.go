package testutils

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/google/go-cmp/cmp"
)

// CompareResult represents the aggregated result of one or more comparisons
// performed during a test.
//
// It collects multiple assertion failures and defers test termination until
// Validate is called, allowing a single test run to report all mismatches
// instead of failing on the first error.
//
// A CompareResult is successful when OK is true. Any call that records a
// difference (e.g., AddDiff, Assert, AssertErr) sets OK to false and appends
// a human-readable message to Diff.
//
// After performing comparisons, Validate must be called to fail the test if
// any mismatches were recorded.
//
// Example:
//
//	main := NewCompareResult()
//
//	create := NewCompareResult()
//	create.Assert("Create", expectedCreate, actualCreate)
//	create.AssertErr("Create.Err", expectedCreateErr, actualCreateErr)
//
//	update := NewCompareResult()
//	update.Assert("Update", expectedUpdate, actualUpdate)
//	update.AssertErr("Update.Err", expectedUpdateErr, actualUpdateErr)
//
//	main.Merge(create)
//	main.Merge(update)
//
//	main.Validate(t, "TestFlow")
type CompareResult struct {
	OK   bool     // True if comparison passed completely, false otherwise
	Diff []string // List of human-readable failure descriptions, empty if OK
}

// NewCompareResult creates a successful comparison result instance.
func NewCompareResult() *CompareResult {
	return &CompareResult{OK: true}
}

// AddDiff marks the comparison as failed and appends a custom failure message.
//
// This is the primary way to report simple failures like row count mismatches
// or pagination URL differences.
func (r *CompareResult) AddDiff(diff string, args ...any) *CompareResult {
	if len(args) != 0 {
		return r.AddDifference(fmt.Sprintf(diff, args...))
	}

	return r.AddDifference(diff)
}

func (r *CompareResult) AddDifference(diff string) *CompareResult {
	r.OK = false
	r.Diff = append(r.Diff, diff)
	return r
}

// Assert compares two data structures using github.com/go-test/deep.Equal
// and records a formatted mismatch report for the specified data name.
// Returns true if mismatch found (and recorded), false if exact match.
//
// Example output:
//
//	Data[0].Fields[stagename] mismatch:
//	❌ Prospecting != PROSPECTING
//
//	Data[0].Raw[OpportunityContactRoles] mismatch:
//	❌ map[totalSize]: 2 != 3
//
// No-op (returns true) if structures match exactly.
func (r *CompareResult) Assert(dataName string, expectedData, gotData any) bool {
	status, isMap := r.tryAssertMap(dataName, expectedData, gotData)
	if isMap {
		return status
	}

	diff := deep.Equal(gotData, expectedData)
	if len(diff) == 0 {
		return true
	}

	list := make([]string, len(diff))
	for index, text := range diff {
		// Tabulated list of mismatches.
		list[index] = fmt.Sprintf("\t❌ %v", text)
	}

	message := fmt.Sprintf("%v mismatch:\n%v", dataName, strings.Join(list, "\n"))

	r.Diff = append(r.Diff, message)
	r.OK = false

	return false
}

func (r *CompareResult) tryAssertMap(dataName string, expectedData, gotData any) (status bool, isMap bool) {
	_, firstIsMap := expectedData.(map[string]any)
	_, secondIsMap := gotData.(map[string]any)
	if !(firstIsMap && secondIsMap) {
		return false, false
	}

	diff := cmp.Diff(expectedData, gotData)
	if diff != "" {
		r.AddDiff("%v: mismatch (-expected +got):\n%s", dataName, diff)
		return false, true
	}

	return true, true
}

// AssertErr compares expected and actual errors and records a mismatch
// in the result if they do not match.
//
// Matching rules:
//
//   - If both expectedErr and actualErr are nil → considered a match.
//   - If only one is nil → mismatch.
//   - Otherwise, errors are compared using errors.Is (semantic match).
//   - If expectedErr is of type StrError, errors are compared as substrings.
//
// Returns true if the errors match, false if a mismatch is found.
func (r *CompareResult) AssertErr(dataName string, expectedErr, actualErr error) bool {
	if expectedErr == nil {
		if actualErr != nil {
			r.AddDiff("%s: expected no error, got: (%v)", dataName, actualErr)
			return false
		}

		// Both errors are nil.
		return true
	}

	if actualErr == nil {
		r.AddDiff("%s: expected error: (%v), got nil", dataName, expectedErr)
		return false
	}

	// Default behavior: strict semantic comparison using errors.Is.
	if !errors.Is(actualErr, expectedErr) {
		// Special marker handling: detect whether the expected error is a StrError marker.
		if _, ok := errors.AsType[StrError](expectedErr); ok {
			// StrError allows flexible comparison:
			// prefer errors.Is, but fall back to message containment.
			if !strings.Contains(actualErr.Error(), expectedErr.Error()) {
				r.AddDiff("%s: expected error: (%v), got: (%v)", dataName, expectedErr, actualErr)
				return false
			}

			// Expected substring is found inside the error.
			return true
		}

		r.AddDiff("%s: expected error: (%v), got: (%v)", dataName, expectedErr, actualErr)
		return false
	}

	// Both errors match.
	return true
}

// Merge combines another CompareResult into the receiver.
//
// Updates OK status (true only if both were true) and concatenates all Diff messages.
// Ignores nil other results. Used for chaining multiple sub-comparisons.
func (r *CompareResult) Merge(other *CompareResult) {
	if other == nil {
		return
	}

	// Success requires both to succeed.
	// Fail if either failed.
	r.OK = r.OK && other.OK
	r.Diff = append(r.Diff, other.Diff...)
}

// Validate finalizes the comparison and fails the test if any mismatches
// were recorded.
//
// If the CompareResult is successful (OK == true), Validate is a no-op.
//
// Otherwise, it builds a structured failure message that includes the provided
// testName and all collected Diff entries, then terminates the test using
// t.Fatal. Each diff is numbered in order of occurrence to improve readability.
//
// This method is intended to be called once at the end of a test after all
// assertions have been executed.
func (r *CompareResult) Validate(t *testing.T, testName string) {
	if r.OK {
		return
	}

	message := fmt.Sprintf("[%s] some expectations were not satisfied:\n", testName)
	for index, text := range r.Diff {
		message += fmt.Sprintf("(%v) %v\n", index+1, text)
	}

	t.Fatal(message)
}
