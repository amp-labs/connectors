// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// TestMissingScopesStillSatisfiesForbidden is the compatibility guarantee that let this
// sentinel be introduced at all. Every existing errors.Is(err, ErrForbidden) check --
// in this repo and in every consumer of it -- must keep holding for a 403 that is now
// reported more precisely.
func TestMissingScopesStillSatisfiesForbidden(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrMissingScopes, ErrForbidden) {
		t.Error("ErrMissingScopes must also satisfy ErrForbidden; it refines that case rather than replacing it")
	}

	if !errors.Is(ErrMissingScopes, ErrMissingScopes) {
		t.Error("ErrMissingScopes must satisfy itself")
	}

	// The relationship is one-way: a plain forbidden is not a scope problem, and treating
	// it as one would send the customer off to reconnect when their admin needs to fix
	// permissions instead.
	if errors.Is(ErrForbidden, ErrMissingScopes) {
		t.Error("ErrForbidden must NOT satisfy ErrMissingScopes")
	}

	// The relationship must survive wrapping, which is how it always arrives in practice.
	wrapped := fmt.Errorf("fetching contacts: %w", ErrMissingScopes)
	if !errors.Is(wrapped, ErrForbidden) || !errors.Is(wrapped, ErrMissingScopes) {
		t.Error("wrapping must preserve both identities")
	}
}

// TestClassOfMissingScopes pins the class itself.
func TestClassOfMissingScopes(t *testing.T) {
	t.Parallel()

	if got := ClassOf(ErrMissingScopes); got != ErrorClassInsufficientScope {
		t.Errorf("ClassOf(ErrMissingScopes) = %q, want %q", got, ErrorClassInsufficientScope)
	}

	// Classification must not be lost to the ErrForbidden that ErrMissingScopes also
	// unwraps to. If it were, the refinement would be invisible exactly where it matters.
	if got := ClassOf(fmt.Errorf("write failed: %w", ErrMissingScopes)); got != ErrorClassInsufficientScope {
		t.Errorf("ClassOf(wrapped ErrMissingScopes) = %q, want %q", got, ErrorClassInsufficientScope)
	}
}

// TestHTTPErrorClassPrefersScopeRefinementOver403 covers the precedence change. Before it,
// classOfHTTPStatus won unconditionally, so a 403 wrapping ErrMissingScopes still reported
// "forbidden" and the new class was unreachable in practice.
func TestHTTPErrorClassPrefersScopeRefinementOver403(t *testing.T) {
	t.Parallel()

	scoped := NewHTTPError(http.StatusForbidden, nil, nil, ErrMissingScopes)
	if got := ClassOf(scoped); got != ErrorClassInsufficientScope {
		t.Errorf("403 wrapping ErrMissingScopes = %q, want %q", got, ErrorClassInsufficientScope)
	}
}

// TestHTTPErrorClassUnchangedForEverythingElse guards the blast radius: the refinement is
// registered only for 403, so no other status changes behaviour.
func TestHTTPErrorClassUnchangedForEverythingElse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		err    error
		want   ErrorClass
	}{
		{"plain 403 is still forbidden", http.StatusForbidden, ErrForbidden, ErrorClassForbidden},
		{"403 with an unrelated sentinel is still forbidden", http.StatusForbidden, ErrRetryable, ErrorClassForbidden},
		{"401 still wins over the wrapped sentinel", http.StatusUnauthorized, ErrMissingScopes, ErrorClassAuthInvalidated},
		{"429 still wins over the wrapped sentinel", http.StatusTooManyRequests, ErrMissingScopes, ErrorClassRateLimited},
		{"500 still wins over the wrapped sentinel", http.StatusInternalServerError, ErrMissingScopes, ErrorClassProvider5xx},
		{"unmapped status still defers to the sentinel", http.StatusTeapot, ErrMissingScopes, ErrorClassInsufficientScope},
		{"unmapped status with no sentinel is unknown", http.StatusTeapot, nil, ErrorClassUnknown},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			httpErr := NewHTTPError(testCase.status, nil, nil, testCase.err)
			if got := ClassOf(httpErr); got != testCase.want {
				t.Errorf("status %d = %q, want %q", testCase.status, got, testCase.want)
			}
		})
	}
}

// TestClassOfMessageInsufficientScope covers the string fallback, which is what runs when
// Temporal or another layer has flattened the typed chain into a bare message.
func TestClassOfMessageInsufficientScope(t *testing.T) {
	t.Parallel()

	cases := []string{
		`HTTP status 403: missing scopes: {"category":"MISSING_SCOPES"}`,
		"forbidden: MISSING_SCOPES",
		`403: error="insufficient_scope"`,
	}

	for _, msg := range cases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			if got := classOfMessage(msg); got != ErrorClassInsufficientScope {
				t.Errorf("classOfMessage(%q) = %q, want %q", msg, got, ErrorClassInsufficientScope)
			}
		})
	}
}

// TestRefinesClass pins the helper directly, including that it is not reflexive and does
// not admit arbitrary pairs.
func TestRefinesClass(t *testing.T) {
	t.Parallel()

	if !refinesClass(ErrorClassForbidden, ErrorClassInsufficientScope) {
		t.Error("insufficient_scope must refine forbidden")
	}

	if refinesClass(ErrorClassForbidden, ErrorClassForbidden) {
		t.Error("a class must not refine itself")
	}

	if refinesClass(ErrorClassForbidden, "") {
		t.Error("an absent class must not refine anything")
	}

	if refinesClass(ErrorClassAuthInvalidated, ErrorClassInsufficientScope) {
		t.Error("insufficient_scope must not refine auth_invalidated")
	}

	if refinesClass(ErrorClassForbidden, ErrorClassRateLimited) {
		t.Error("unregistered pairs must not refine")
	}
}

// TestClassOfMessageTableIsOrderSensitive guards the switch-to-table conversion. In a
// switch the first matching case wins by construction; in a table that is a property of
// the data, so it needs a test. This message matches both the permanent-5xx patterns and
// the generic 5xx ones, and must still resolve to the more specific class.
func TestClassOfMessageTableIsOrderSensitive(t *testing.T) {
	t.Parallel()

	msg := "HTTP status 503: non-retryable server error"
	if got := classOfMessage(msg); got != ErrorClassProvider5xxPermanent {
		t.Errorf("classOfMessage(%q) = %q, want %q", msg, got, ErrorClassProvider5xxPermanent)
	}
}

// TestClassOfMessagePatternsNotCoveredElsewhere fills the gaps left by
// TestClassOf_StringFallbacks, so the whole table is exercised after the conversion.
func TestClassOfMessagePatternsNotCoveredElsewhere(t *testing.T) {
	t.Parallel()

	cases := []struct {
		msg  string
		want ErrorClass
	}{
		{"access token invalid", ErrorClassAuthInvalidated},
		{"the credentials have been marked as invalid", ErrorClassAuthInvalidated},
		{"Could not find a property named 'foo'", ErrorClassSchemaDriftObject},
		{"feature_disabled", ErrorClassAPIDisabled},
		{"missing secret", ErrorClassBadRequest},
		{"non-retryable server error", ErrorClassProvider5xxPermanent},
	}

	for _, testCase := range cases {
		t.Run(testCase.msg, func(t *testing.T) {
			t.Parallel()

			if got := classOfMessage(testCase.msg); got != testCase.want {
				t.Errorf("classOfMessage(%q) = %q, want %q", testCase.msg, got, testCase.want)
			}
		})
	}
}
