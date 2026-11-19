package testconn

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// RequireNoFailures fails the test if the batch result reports any failed id.
// Batch connectors report every requested id absent from Rows, so this also catches
// rows whose Id does not match the id that was requested.
func RequireNoFailures(t *testing.T, result *common.BatchReadResult) {
	t.Helper()

	if result.HasFailures() {
		t.Errorf("expected no failures, got %+v", result.Failures)
	}
}

// RequireRowIds checks that the batch result's rows carry exactly wantIds, in order.
func RequireRowIds(t *testing.T, result *common.BatchReadResult, wantIds ...string) {
	t.Helper()

	if result == nil {
		t.Fatal("expected a result, got nil")
	}

	gotIds := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		gotIds = append(gotIds, row.Id)
	}

	if !slices.Equal(gotIds, wantIds) {
		t.Errorf("row ids %v, want %v", gotIds, wantIds)
	}
}

// RequireRowIdsAnyOrder checks that the batch result's rows carry exactly wantIds, in any order. Use it
// for connectors that add rows in the order their fetches finish.
func RequireRowIdsAnyOrder(t *testing.T, result *common.BatchReadResult, wantIds ...string) {
	t.Helper()

	if result == nil {
		t.Fatal("expected a result, got nil")
	}

	gotIds := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		gotIds = append(gotIds, row.Id)
	}

	slices.Sort(gotIds)

	wantSorted := slices.Sorted(slices.Values(wantIds))

	if !slices.Equal(gotIds, wantSorted) {
		t.Errorf("row ids %v, want %v in any order", gotIds, wantSorted)
	}
}

// SortRowsByID sorts rows by id, so a test can index the rows of a connector that adds them in the
// order its fetches finish.
func SortRowsByID(rows []common.ReadResultRow) {
	slices.SortFunc(rows, func(a, b common.ReadResultRow) int {
		return strings.Compare(a.Id, b.Id)
	})
}

// RequireMissingIds checks that the batch result reports exactly wantIds, in requested
// order, as not-found failures wrapping common.ErrNotFound.
func RequireMissingIds(t *testing.T, result *common.BatchReadResult, wantIds ...string) {
	t.Helper()

	if result == nil {
		t.Fatal("expected a result, got nil")
	}

	gotIds := make([]string, 0, len(result.Failures))

	for _, failure := range result.Failures {
		gotIds = append(gotIds, failure.RecordId)

		if failure.Reason != common.FailureReasonNotFound {
			t.Errorf("failure for %q: reason %q, want %q",
				failure.RecordId, failure.Reason, common.FailureReasonNotFound)
		}

		if !errors.Is(failure.Err, common.ErrNotFound) {
			t.Errorf("failure for %q: error %v does not wrap common.ErrNotFound", failure.RecordId, failure.Err)
		}
	}

	if !slices.Equal(gotIds, wantIds) {
		t.Errorf("failed ids %v, want %v", gotIds, wantIds)
	}
}
