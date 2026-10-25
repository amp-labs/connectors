package core

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// TestInterpretJSONError_Forbidden covers the distinction this file now draws: HubSpot
// labels a scope failure in its error body, and the remedy for that ("reconnect granting
// the scope") is the opposite of the remedy for a plain 403 ("change the user's HubSpot
// permissions"). Collapsing both into ErrForbidden made that advice a coin flip.
func TestInterpretJSONError_Forbidden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		wantClass      common.ErrorClass
		wantScopeError bool
	}{
		{
			name: "MISSING_SCOPES category is a scope problem",
			body: `{"status":"error","message":"This app hasn't been granted all required scopes.",` +
				`"correlationId":"a1","category":"MISSING_SCOPES"}`,
			wantClass:      common.ErrorClassInsufficientScope,
			wantScopeError: true,
		},
		{
			name: "BANNED category is not",
			body: `{"status":"error","message":"The app is not allowed to access this portal.",` +
				`"correlationId":"a2","category":"BANNED"}`,
			wantClass:      common.ErrorClassForbidden,
			wantScopeError: false,
		},
		{
			name:           "no category at all is not",
			body:           `{"status":"error","message":"nope","correlationId":"a3"}`,
			wantClass:      common.ErrorClassForbidden,
			wantScopeError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{StatusCode: http.StatusForbidden, Header: http.Header{}}

			err := InterpretJSONError(res, []byte(testCase.body))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Every 403 remains a forbidden error, whichever kind it is. Consumers
			// matching on that must not have to learn about the new sentinel.
			if !errors.Is(err, common.ErrForbidden) {
				t.Error("every 403 must still satisfy errors.Is(err, ErrForbidden)")
			}

			if got := errors.Is(err, common.ErrMissingScopes); got != testCase.wantScopeError {
				t.Errorf("errors.Is(err, ErrMissingScopes) = %v, want %v", got, testCase.wantScopeError)
			}

			if got := common.ClassOf(err); got != testCase.wantClass {
				t.Errorf("ClassOf = %q, want %q", got, testCase.wantClass)
			}

			var httpErr *common.HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("expected *common.HTTPError in chain, got %T", err)
			}

			if string(httpErr.Body) != testCase.body {
				t.Errorf("response body must be preserved for triage\n  want: %s\n  got:  %s",
					testCase.body, httpErr.Body)
			}
		})
	}
}
