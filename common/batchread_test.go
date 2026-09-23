// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"net/http"
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

// A 404 must win over the ErrorClass the error would otherwise resolve to.
// Without the not-found check first, a 404 carrying a bad-request sentinel
// would classify as permanent and lose the "the record is simply gone" signal
// that callers use to drop an id instead of retrying it.
func TestFailureReasonOf_NotFoundBeatsErrorClass(t *testing.T) {
	t.Parallel()

	// A 404 whose wrapped sentinel would otherwise classify as permanent.
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

	// Rows are untouched by a failure being recorded.
	if len(result.Rows) != 1 {
		t.Errorf("len(Rows) = %d, want 1", len(result.Rows))
	}
}

// HasFailures is called on results that callers may not have checked for nil.
func TestBatchReadResult_HasFailuresNil(t *testing.T) {
	t.Parallel()

	var result *BatchReadResult

	if result.HasFailures() {
		t.Error("nil result must report no failures")
	}
}
