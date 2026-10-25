// nolint:revive,godoclint
package common

import (
	"errors"
	"net/http"
	"testing"
)

// TestInterpretErrorHonoursRFC6750Challenge covers the provider-agnostic path. RFC 6750
// section 3.1 gives providers a standard way to say "this token lacks the scope" rather
// than "this user lacks permission"; honouring it means any provider that follows the spec
// gets the distinction without a bespoke interpreter.
func TestInterpretErrorHonoursRFC6750Challenge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		challenge      string
		wantClass      ErrorClass
		wantScopeError bool
	}{
		{
			name:           "insufficient_scope challenge",
			challenge:      `Bearer realm="example", error="insufficient_scope", scope="contacts.read"`,
			wantClass:      ErrorClassInsufficientScope,
			wantScopeError: true,
		},
		{
			name:           "challenge casing varies across providers",
			challenge:      `BEARER REALM="example", ERROR="insufficient_scope"`,
			wantClass:      ErrorClassInsufficientScope,
			wantScopeError: true,
		},
		{
			name:           "a different challenge is not a scope problem",
			challenge:      `Bearer realm="example", error="invalid_token"`,
			wantClass:      ErrorClassForbidden,
			wantScopeError: false,
		},
		{
			name:           "no challenge at all",
			challenge:      "",
			wantClass:      ErrorClassForbidden,
			wantScopeError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{StatusCode: http.StatusForbidden, Header: http.Header{}}
			if testCase.challenge != "" {
				res.Header.Set("WWW-Authenticate", testCase.challenge)
			}

			err := InterpretError(res, []byte(`{"error":"nope"}`))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Whichever kind it is, it is still a 403.
			if !errors.Is(err, ErrForbidden) {
				t.Error("every 403 must still satisfy errors.Is(err, ErrForbidden)")
			}

			if got := errors.Is(err, ErrMissingScopes); got != testCase.wantScopeError {
				t.Errorf("errors.Is(err, ErrMissingScopes) = %v, want %v", got, testCase.wantScopeError)
			}

			if got := ClassOf(err); got != testCase.wantClass {
				t.Errorf("ClassOf = %q, want %q", got, testCase.wantClass)
			}
		})
	}
}

// TestInterpretErrorChallengeIgnoredOnOtherStatuses: the challenge only reclassifies a 403.
// A 401 is an invalidated credential whatever else the provider says, and that has its own
// remedy (refresh or reconnect) which the retry machinery already acts on.
func TestInterpretErrorChallengeIgnoredOnOtherStatuses(t *testing.T) {
	t.Parallel()

	res := &http.Response{StatusCode: http.StatusUnauthorized, Header: http.Header{}}
	res.Header.Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)

	err := InterpretError(res, nil)
	if !errors.Is(err, ErrAccessToken) {
		t.Error("a 401 must still be an access-token error")
	}

	if got := ClassOf(err); got != ErrorClassAuthInvalidated {
		t.Errorf("ClassOf = %q, want %q", got, ErrorClassAuthInvalidated)
	}
}
