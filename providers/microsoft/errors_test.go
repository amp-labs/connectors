package microsoft

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestHandleErrorResponse(t *testing.T) { //nolint:funlen
	t.Parallel()

	const (
		genericInvalidTokenBody = `{"error":{"code":"InvalidAuthenticationToken","message":"Access token validation failure."}}`
		expiredTokenBody        = `{"error":{"code":"InvalidAuthenticationToken","message":"Lifetime validation failed, the token is expired."}}`
		notYetValidTokenBody    = `{"error":{"code":"InvalidAuthenticationToken","message":"The token is not yet valid."}}`
		accessExpiredBody       = `{"error":{"code":"InvalidAuthenticationToken","message":"Access token has expired or is not yet valid and can no longer be used."}}`
	)

	tests := []struct {
		name             string
		status           int
		wwwAuthenticate  string
		body             string
		wantRetryable    bool
		wantAccessToken  bool
		wantErrSubstring string
	}{
		{
			name:            "401 with CAE insufficient_claims is bad-creds (needs re-auth, not retryable)",
			status:          http.StatusUnauthorized,
			wwwAuthenticate: `Bearer realm="", authorization_uri="https://login.microsoftonline.com/common/oauth2/authorize", error="insufficient_claims", claims="eyJhY2Nlc3NfdG9rZW4iOnsiYWNycyI6eyJlc3NlbnRpYWwiOnRydWUsInZhbHVlIjoiYzEifX19"`,
			body:            genericInvalidTokenBody,
			wantAccessToken: true,
		},
		{
			name:            "401 with interaction_required is bad-creds (needs user interaction, not retryable)",
			status:          http.StatusUnauthorized,
			wwwAuthenticate: `Bearer error="interaction_required", error_description="additional auth required"`,
			body:            genericInvalidTokenBody,
			wantAccessToken: true,
		},
		{
			name:             "401 InvalidAuthenticationToken lifetime-expired is retryable",
			status:           http.StatusUnauthorized,
			body:             expiredTokenBody,
			wantRetryable:    true,
			wantErrSubstring: "Lifetime validation failed",
		},
		{
			name:          "401 InvalidAuthenticationToken not-yet-valid is retryable",
			status:        http.StatusUnauthorized,
			body:          notYetValidTokenBody,
			wantRetryable: true,
		},
		{
			name:          "401 InvalidAuthenticationToken access-expired is retryable",
			status:        http.StatusUnauthorized,
			body:          accessExpiredBody,
			wantRetryable: true,
		},
		{
			name:            "401 with no CAE header and non-transient message falls through to access-token error",
			status:          http.StatusUnauthorized,
			body:            genericInvalidTokenBody,
			wantAccessToken: true,
		},
		{
			name:            "401 with empty body falls through to access-token error",
			status:          http.StatusUnauthorized,
			body:            "",
			wantAccessToken: true,
		},
		{
			name:             "403 is unchanged (forbidden)",
			status:           http.StatusForbidden,
			body:             `{"error":{"code":"Forbidden","message":"Access denied."}}`,
			wantErrSubstring: "forbidden",
		},
		{
			name:             "400 is unchanged (bad request)",
			status:           http.StatusBadRequest,
			body:             `{"error":{"code":"BadRequest","message":"bad input"}}`,
			wantErrSubstring: "bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{
				StatusCode: tt.status,
				Header:     http.Header{},
			}
			if tt.wwwAuthenticate != "" {
				res.Header.Set("WWW-Authenticate", tt.wwwAuthenticate)
			}

			err := handleErrorResponse(res, []byte(tt.body))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantRetryable {
				if !errors.Is(err, common.ErrRetryable) {
					t.Errorf("expected errors.Is(err, ErrRetryable), got %v", err)
				}
				if errors.Is(err, common.ErrAccessToken) {
					t.Errorf("did not expect ErrAccessToken, got %v", err)
				}
			}

			if tt.wantAccessToken {
				if !errors.Is(err, common.ErrAccessToken) {
					t.Errorf("expected errors.Is(err, ErrAccessToken), got %v", err)
				}
				if errors.Is(err, common.ErrRetryable) {
					t.Errorf("did not expect ErrRetryable, got %v", err)
				}
			}

			if tt.wantErrSubstring != "" && !strings.Contains(err.Error(), tt.wantErrSubstring) {
				t.Errorf("expected error to contain %q, got %v", tt.wantErrSubstring, err)
			}
		})
	}
}

func TestClaimsChallengeErrorExposesStructuredFields(t *testing.T) {
	t.Parallel()

	reqURL, err := url.Parse("https://graph.microsoft.com/v1.0/users")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    reqURL,
	}

	res := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Header:     http.Header{},
		Request:    req,
	}
	res.Header.Set("WWW-Authenticate",
		`Bearer error="insufficient_claims", claims="eyJhY2Nlc3NfdG9rZW4iOnsiYWNycyI6eyJlc3NlbnRpYWwiOnRydWUsInZhbHVlIjoiYzEifX19"`)
	res.Header.Set("x-ms-request-id", "e9e89775-0174-4796-8446-a8fc421b3600")

	returned := handleErrorResponse(res, []byte(`{"error":{"code":"InvalidAuthenticationToken","message":"x"}}`))

	// Claim challenges are NOT retryable — retries cannot satisfy the
	// challenge without user action. They must surface as bad credentials
	// (common.ErrAccessToken) so the server flips the connection and the
	// user is prompted to re-auth.
	if !errors.Is(returned, common.ErrAccessToken) {
		t.Fatalf("expected errors.Is(err, ErrAccessToken) to be true, got %v", returned)
	}

	if errors.Is(returned, common.ErrRetryable) {
		t.Fatalf("claim challenge must not be retryable, but errors.Is(err, ErrRetryable) was true: %v", returned)
	}

	var chErr *ClaimsChallengeError
	if !errors.As(returned, &chErr) {
		t.Fatalf("expected errors.As to populate *ClaimsChallengeError, got %T: %v", returned, returned)
	}

	if chErr.Reason != "insufficient_claims" {
		t.Errorf("Reason: got %q, want %q", chErr.Reason, "insufficient_claims")
	}

	if chErr.RequestMethod != http.MethodGet {
		t.Errorf("RequestMethod: got %q, want %q", chErr.RequestMethod, http.MethodGet)
	}

	if chErr.RequestURL != "https://graph.microsoft.com/v1.0/users" {
		t.Errorf("RequestURL: got %q", chErr.RequestURL)
	}

	if chErr.MSRequestID != "e9e89775-0174-4796-8446-a8fc421b3600" {
		t.Errorf("MSRequestID: got %q", chErr.MSRequestID)
	}

	if !strings.Contains(chErr.WWWAuthenticate, `error="insufficient_claims"`) {
		t.Errorf("WWWAuthenticate: missing original header content, got %q", chErr.WWWAuthenticate)
	}

	// Error() should include every diagnostic field so the existing server
	// logger pattern (logger.Error("...", "error", err)) surfaces them all
	// without requiring an errors.As block on the server side.
	msg := chErr.Error()
	for _, want := range []string{
		"microsoft claims challenge",
		"insufficient_claims",
		http.MethodGet,
		"https://graph.microsoft.com/v1.0/users",
		"x-ms-request-id=e9e89775-0174-4796-8446-a8fc421b3600",
		"www-authenticate=",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() missing %q\ngot: %s", want, msg)
		}
	}
}

func TestClaimsChallengeReason(t *testing.T) {
	t.Parallel()

	cases := []struct {
		header string
		want   string
		ok     bool
	}{
		{"", "", false},
		{`Bearer realm=""`, "", false},
		{`Bearer error="insufficient_claims", claims="abc"`, "insufficient_claims", true},
		{`Bearer ERROR="INSUFFICIENT_CLAIMS"`, "insufficient_claims", true},
		{`Bearer error="interaction_required"`, "interaction_required", true},
		{`Bearer error="invalid_token"`, "", false},
	}

	for _, c := range cases {
		got, ok := claimsChallengeReason(c.header)
		if ok != c.ok || got != c.want {
			t.Errorf("claimsChallengeReason(%q) = (%q,%v), want (%q,%v)", c.header, got, ok, c.want, c.ok)
		}
	}
}
