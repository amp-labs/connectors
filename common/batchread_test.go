// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"
)

func TestFailureReasonOf_Nil(t *testing.T) {
	t.Parallel()

	if got := FailureReasonOf(nil); got != FailureReasonUnknown {
		t.Errorf("FailureReasonOf(nil) = %q, want %q", got, FailureReasonUnknown)
	}
}

func TestFailureReasonOf_NotFound(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
	}{
		{"sentinel", ErrNotFound},
		{"wrapped sentinel", fmt.Errorf("fetch job-1: %w", ErrNotFound)},
		{"http 404", NewHTTPError(http.StatusNotFound, nil, nil, ErrNotFound)},
		{"http 404 joined with context", errors.Join(
			errors.New("fetch job-1"), //nolint:err113
			NewHTTPError(http.StatusNotFound, nil, nil, ErrNotFound),
		)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := FailureReasonOf(tc.err); got != FailureReasonNotFound {
				t.Errorf("FailureReasonOf(%v) = %q, want %q", tc.err, got, FailureReasonNotFound)
			}
		})
	}
}

// A 404 is classified not-found even when its ErrorClass says otherwise.
func TestFailureReasonOf_NotFoundBeatsErrorClass(t *testing.T) {
	t.Parallel()

	err := NewHTTPError(http.StatusNotFound, nil, nil, ErrBadRequest)

	if ClassOf(err) != ErrorClassBadRequest {
		t.Fatalf("precondition: ClassOf = %q, want %q", ClassOf(err), ErrorClassBadRequest)
	}

	if got := FailureReasonOf(err); got != FailureReasonNotFound {
		t.Errorf("FailureReasonOf(%v) = %q, want %q", err, got, FailureReasonNotFound)
	}
}

func TestFailureReasonOf_ByErrorClass(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want FailureReason
	}{
		{ErrLimitExceeded, FailureReasonTransient},
		{ErrServer, FailureReasonTransient},
		{ErrRetryable, FailureReasonTransient},
		{ErrAccessToken, FailureReasonPermanent},
		{ErrInvalidGrant, FailureReasonPermanent},
		{ErrForbidden, FailureReasonPermanent},
		{ErrApiDisabled, FailureReasonPermanent},
		{ErrBadRequest, FailureReasonPermanent},
		{ErrCursorGone, FailureReasonPermanent},
		{errors.New("something nobody classified"), FailureReasonUnknown}, //nolint:err113
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			t.Parallel()

			if got := FailureReasonOf(tc.err); got != tc.want {
				t.Errorf("FailureReasonOf(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

func TestNewRecordFailure(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("fetch contact-9: %w", ErrNotFound)

	failure := NewRecordFailure("contact-9", cause)

	if failure.RecordId != "contact-9" {
		t.Errorf("RecordId = %q, want %q", failure.RecordId, "contact-9")
	}

	if !errors.Is(failure.Err, ErrNotFound) {
		t.Errorf("Err = %v, want it to wrap ErrNotFound", failure.Err)
	}

	if failure.Reason != FailureReasonNotFound {
		t.Errorf("Reason = %q, want %q", failure.Reason, FailureReasonNotFound)
	}
}

func TestBatchReadResult_AddFailure(t *testing.T) {
	t.Parallel()

	result := NewBatchReadResult([]ReadResultRow{{Id: "1"}})

	if result.HasFailures() {
		t.Error("a fresh result built from rows must have no failures")
	}

	result.AddFailure("2", ErrLimitExceeded)

	if !result.HasFailures() {
		t.Fatal("HasFailures() = false after AddFailure")
	}

	if len(result.Failures) != 1 {
		t.Fatalf("len(Failures) = %d, want 1", len(result.Failures))
	}

	if result.Failures[0].Reason != FailureReasonTransient {
		t.Errorf("Reason = %q, want %q", result.Failures[0].Reason, FailureReasonTransient)
	}

	if len(result.Rows) != 1 {
		t.Errorf("len(Rows) = %d, want 1", len(result.Rows))
	}
}

// HasFailures handles a nil result.
func TestBatchReadResult_HasFailuresNil(t *testing.T) {
	t.Parallel()

	var result *BatchReadResult

	if result.HasFailures() {
		t.Error("nil result must report no failures")
	}
}

// Not-found errors become failures in request order; every other error is returned, joined in
// request order and naming its id.
func TestBatchReadResult_ReportFetchErrors(t *testing.T) {
	t.Parallel()

	ids := []string{"a", "b", "c", "d", "e"}
	serverErr := NewHTTPError(http.StatusInternalServerError, nil, nil, ErrServer)

	result := NewBatchReadResult(nil)

	err := result.ReportFetchErrors(ids, map[int]error{
		4: NewHTTPError(http.StatusNotFound, nil, nil, ErrRetryable),
		3: ErrLimitExceeded,
		1: fmt.Errorf("wrapped: %w", ErrNotFound),
		0: serverErr,
	})

	got := make([]string, 0, len(result.Failures))
	for _, failure := range result.Failures {
		got = append(got, failure.RecordId)

		if failure.Reason != FailureReasonNotFound {
			t.Errorf("failure for %q: reason %q, want %q", failure.RecordId, failure.Reason, FailureReasonNotFound)
		}
	}

	if want := []string{"b", "e"}; !slices.Equal(got, want) {
		t.Errorf("failures %v, want %v", got, want)
	}

	if !errors.Is(err, ErrServer) || !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("error %v must wrap every non-not-found error", err)
	}

	if errors.Is(err, ErrNotFound) {
		t.Errorf("error %v must not include the not-found errors", err)
	}

	if want := "record a: " + serverErr.Error() + "\nrecord d: " + ErrLimitExceeded.Error(); err.Error() != want {
		t.Errorf("error %q, want %q", err.Error(), want)
	}
}

// No errors, or only not-found errors, leave the call successful.
func TestBatchReadResult_ReportFetchErrorsNoError(t *testing.T) {
	t.Parallel()

	result := NewBatchReadResult(nil)

	if err := result.ReportFetchErrors([]string{"a"}, nil); err != nil {
		t.Errorf("no errors: got %v", err)
	}

	if err := result.ReportFetchErrors([]string{"a"}, map[int]error{0: ErrNotFound}); err != nil {
		t.Errorf("only not-found: got %v", err)
	}

	if len(result.Failures) != 1 {
		t.Errorf("expected one not-found failure, got %d", len(result.Failures))
	}
}
