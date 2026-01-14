package mockcond

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestPermuteQueryParam(t *testing.T) {
	// --- input fields we want permutations for ---
	fields := []string{"A", "B", "C"}

	// --- expected permutations ---
	// This must match your generatePermutations logic exactly.
	expected := []string{
		"A,B,C",
		"B,A,C",
		"C,B,A",
		"B,C,A",
		"A,C,B",
		"C,A,B",
	}

	expectedSet := make(map[string]bool)
	for _, s := range expected {
		expectedSet[s] = true
	}

	// --- build the Or(...) using Permute ---
	conditions := Permute(
		func(order []string) Condition {
			selector := strings.Join(order, ",")

			return QueryParam("q", "SELECT "+selector+" FROM opportunity")
		}, fields...)

	// Sanity check: count match
	if len(conditions) != len(expected) {
		t.Fatalf("expected %d conditions but Permute created %d",
			len(expected), len(conditions))
	}

	// 1. Ensure each generated permutation is expected.
	// There shouldn't be more conditions than expected.
	for _, cond := range conditions {
		// Try all expected requests until one matches
		matched := false

		for _, selectExpectation := range expected {
			r := newRequestWithQuery(t, "SELECT "+selectExpectation+" FROM opportunity")
			if cond.EvaluateCondition(nil, r) {
				matched = true
				break
			}
		}

		if !matched {
			t.Fatalf("generated Condition did not match any expected permutation")
		}
	}

	// 2. Ensure Or(...) matches *every* expected permutation.
	for _, selectExpectation := range expected {
		r := newRequestWithQuery(t, "SELECT "+selectExpectation+" FROM opportunity")
		if !conditions.EvaluateCondition(nil, r) {
			t.Fatalf("Or(...) should match permutation %q but did not", selectExpectation)
		}
	}

	// --- 3. Ensure Or(...) rejects queries NOT in the permutations ---
	invalidValue := "SELECT A,C,D FROM opportunity"
	invalidReq := newRequestWithQuery(t, invalidValue)

	if conditions.EvaluateCondition(nil, invalidReq) {
		t.Fatalf("Or(...) matched an invalid query %q", invalidValue)
	}
}

// helper: create a request with the given ?q=... parameter
func newRequestWithQuery(t *testing.T, queryValue string) *http.Request {
	t.Helper()

	r, err := http.NewRequest("GET", "/test?q="+url.QueryEscape(queryValue), nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	return r
}
