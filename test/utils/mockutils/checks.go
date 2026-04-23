package mockutils

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

// CompareResult holds the result of a comparison operation between actual and expected values.
// It tracks whether the comparison passed (OK) and collects detailed failure messages (Diff)
// for precise test failure diagnostics.
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
func (r *CompareResult) AddDiff(diff string) {
	r.OK = false
	r.Diff = append(r.Diff, diff)
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
// No-op (returns false) if structures match exactly.
func (r *CompareResult) Assert(dataName string, expectedData, gotData any) bool {
	diff := deep.Equal(gotData, expectedData)
	if len(diff) == 0 {
		return false
	}

	list := make([]string, len(diff))
	for index, text := range diff {
		// Tabulated list of mismatches.
		list[index] = fmt.Sprintf("\t❌ %v", text)
	}

	message := fmt.Sprintf("%v mismatch:\n%v", dataName, strings.Join(list, "\n"))

	r.Diff = append(r.Diff, message)
	r.OK = false

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

func DoesObjectCorrespondToString(object any, correspondent string) bool {
	if object == nil && len(correspondent) == 0 {
		return true
	}

	switch object.(type) {
	case float64:
		f, err := strconv.ParseFloat(correspondent, 10)
		if err != nil {
			return false
		}

		return f == object
	}

	return fmt.Sprintf("%v", object) == correspondent
}

func NoErrors(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("failed to start test, %v", err)
	}
}
