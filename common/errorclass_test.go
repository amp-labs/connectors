// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestClassOf_Nil(t *testing.T) {
	t.Parallel()

	if got := ClassOf(nil); got != ErrorClassNone {
		t.Errorf("ClassOf(nil) = %q, want %q", got, ErrorClassNone)
	}
}

func TestClassOf_Sentinels(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want ErrorClass
	}{
		{ErrAccessToken, ErrorClassAuthInvalidated},
		{ErrInvalidGrant, ErrorClassAuthInvalidated},
		{ErrForbidden, ErrorClassForbidden},
		{ErrApiDisabled, ErrorClassAPIDisabled},
		{ErrCursorGone, ErrorClassCursorGone},
		{ErrLimitExceeded, ErrorClassRateLimited},
		{ErrResultsLimitExceeded, ErrorClassRateLimited},
		{ErrBadRequest, ErrorClassBadRequest},
		{ErrCaller, ErrorClassBadRequest},
		{ErrServer, ErrorClassProvider5xx},
		{ErrRetryable, ErrorClassRetryable},
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			t.Parallel()

			if got := ClassOf(tc.err); got != tc.want {
				t.Errorf("ClassOf(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

func TestClassOf_SentinelWrapped(t *testing.T) {
	t.Parallel()

	// fmt.Errorf(%w) wrapping must preserve classification via the chain.
	wrapped := fmt.Errorf("context: %w", ErrAccessToken)
	if got := ClassOf(wrapped); got != ErrorClassAuthInvalidated {
		t.Errorf("ClassOf(wrapped ErrAccessToken) = %q, want %q", got, ErrorClassAuthInvalidated)
	}
}

func TestClassOf_HTTPErrorByStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		want   ErrorClass
	}{
		{"401 unauthorized", http.StatusUnauthorized, ErrorClassAuthInvalidated},
		{"403 forbidden", http.StatusForbidden, ErrorClassForbidden},
		{"429 rate limited", http.StatusTooManyRequests, ErrorClassRateLimited},
		{"477 hubspot migration", 477, ErrorClassProviderMigration},
		{"500 server error", http.StatusInternalServerError, ErrorClassProvider5xx},
		{"502 bad gateway", http.StatusBadGateway, ErrorClassProvider5xx},
		{"503 unavailable", http.StatusServiceUnavailable, ErrorClassProvider5xx},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Wrap a generic error; status code should drive classification,
			// not the wrapped sentinel.
			httpErr := NewHTTPError(tc.status, nil, nil, errors.New("whatever"))
			if got := ClassOf(httpErr); got != tc.want {
				t.Errorf("ClassOf(HTTPError{status=%d}) = %q, want %q", tc.status, got, tc.want)
			}
		})
	}
}

func TestClassOf_HTTPErrorFallsThroughToWrappedSentinel(t *testing.T) {
	t.Parallel()

	// A 400 status has no direct mapping — we should consult the wrapped
	// sentinel. ErrCaller → bad_request.
	httpErr := NewHTTPError(http.StatusBadRequest, nil, nil, ErrCaller)
	if got := ClassOf(httpErr); got != ErrorClassBadRequest {
		t.Errorf("ClassOf(400 + ErrCaller) = %q, want %q", got, ErrorClassBadRequest)
	}
}

func TestClassOf_StringFallbacks(t *testing.T) {
	t.Parallel()

	// These errors arrive with no typed sentinel we can see — either because
	// Temporal stringified the chain, or because they come from a provider
	// response body. The string fallback catches them.
	cases := []struct {
		name string
		msg  string
		want ErrorClass
	}{
		{
			name: "salesforce missing custom field",
			msg:  `bad request: No such column 'signed_mnda__c' on entity 'Opportunity'`,
			want: ErrorClassSchemaDriftField,
		},
		{
			name: "salesforce bad relationship",
			msg:  `Didn't understand relationship 'custom_fields' in field path`,
			want: ErrorClassSchemaDriftField,
		},
		{
			name: "dynamics missing object",
			msg:  `bad request: not found: Resource not found for the segment 'systemuser'`,
			want: ErrorClassSchemaDriftObject,
		},
		{
			name: "hubspot migration (via 477 body, but here as raw string)",
			msg:  `Hub 12345678 is currently being migrated between data hosting locations`,
			want: ErrorClassProviderMigration,
		},
		{
			name: "outreach locked user",
			msg:  `This User Is Locked`,
			want: ErrorClassAuthInvalidated,
		},
		{
			name: "google calendar api disabled",
			msg:  `Google Calendar API has not been used in project 123 before`,
			want: ErrorClassAPIDisabled,
		},
		{
			name: "gong cursor expired",
			msg:  `cursor has expired`,
			want: ErrorClassCursorGone,
		},
		{
			name: "salesforce request-line too large",
			msg:  `HTTP ERROR 431 Request Header Fields Too Large`,
			want: ErrorClassBadRequest,
		},
		{
			name: "google precondition failed",
			msg:  `Precondition check failed`,
			want: ErrorClassBadRequest,
		},
		{
			name: "claricopilot 502",
			msg:  `HTTP status 502: internal server error`,
			want: ErrorClassProvider5xx,
		},
		{
			name: "salesloft unexpected server error",
			msg:  `Unexpected server error occurred. Please try again.`,
			want: ErrorClassProvider5xx,
		},
		{
			name: "genuinely unknown message",
			msg:  `something totally different`,
			want: ErrorClassUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ClassOf(errors.New(tc.msg)); got != tc.want {
				t.Errorf("ClassOf(%q) = %q, want %q", tc.msg, got, tc.want)
			}
		})
	}
}

// Verify a custom error type can implement Classifier and wins over sentinels.
type customClassifierErr struct{}

func (customClassifierErr) Error() string          { return "custom" }
func (customClassifierErr) ErrorClass() ErrorClass { return ErrorClassProviderMigration }

func TestClassOf_ClassifierInterfaceWins(t *testing.T) {
	t.Parallel()

	// Wrap a sentinel that would otherwise classify as auth_invalidated.
	// The outer custom error implements Classifier and must win.
	wrapped := fmt.Errorf("%w: %w", customClassifierErr{}, ErrAccessToken)
	if got := ClassOf(wrapped); got != ErrorClassProviderMigration {
		t.Errorf("Classifier did not win over sentinel: got %q, want %q", got, ErrorClassProviderMigration)
	}
}

func TestSentinelErrorsAs(t *testing.T) {
	t.Parallel()

	// Verify errors.As finds the Classifier interface through wrapping.
	wrapped := fmt.Errorf("provider returned: %w", ErrAccessToken)

	var c Classifier
	if !errors.As(wrapped, &c) {
		t.Fatal("errors.As(wrapped, &Classifier) = false, want true")
	}

	if got := c.ErrorClass(); got != ErrorClassAuthInvalidated {
		t.Errorf("ErrorClass() = %q, want %q", got, ErrorClassAuthInvalidated)
	}

	// Verify the inner errors.New is also reachable via Unwrap chain.
	var ce *classedError
	if !errors.As(wrapped, &ce) {
		t.Fatal("errors.As(wrapped, &classedError) = false, want true")
	}

	inner := ce.Unwrap()
	if inner == nil {
		t.Fatal("classedError.Unwrap() = nil, want the inner errors.New")
	}

	if inner.Error() != "access token invalid" {
		t.Errorf("inner error = %q, want %q", inner.Error(), "access token invalid")
	}
}

// TestSentinelErrorsIsStillWorks guards the backward-compat promise that
// callers doing `errors.Is(err, ErrAccessToken)` still behave exactly the
// same after the sentinel migration to newClassedErr.
func TestSentinelErrorsIsStillWorks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		sentinel error
	}{
		{"ErrAccessToken", ErrAccessToken},
		{"ErrForbidden", ErrForbidden},
		{"ErrApiDisabled", ErrApiDisabled},
		{"ErrCursorGone", ErrCursorGone},
		{"ErrLimitExceeded", ErrLimitExceeded},
		{"ErrBadRequest", ErrBadRequest},
		{"ErrCaller", ErrCaller},
		{"ErrServer", ErrServer},
		{"ErrRetryable", ErrRetryable},
		{"ErrInvalidGrant", ErrInvalidGrant},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Wrapped via fmt.Errorf(%w) — the most common pattern.
			wrapped := fmt.Errorf("context: %w", tc.sentinel)
			if !errors.Is(wrapped, tc.sentinel) {
				t.Errorf("errors.Is(wrapped, %s) = false, want true", tc.name)
			}

			// Direct equality — also the most common pattern for the sentinel itself.
			if !errors.Is(tc.sentinel, tc.sentinel) {
				t.Errorf("errors.Is(%s, %s) = false, want true", tc.name, tc.name)
			}
		})
	}
}
