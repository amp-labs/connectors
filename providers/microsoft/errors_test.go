package microsoft

import (
	"errors"
	"net/http"
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
			name:            "401 with CAE insufficient_claims is retryable",
			status:          http.StatusUnauthorized,
			wwwAuthenticate: `Bearer realm="", authorization_uri="https://login.microsoftonline.com/common/oauth2/authorize", error="insufficient_claims", claims="eyJhY2Nlc3NfdG9rZW4iOnsiYWNycyI6eyJlc3NlbnRpYWwiOnRydWUsInZhbHVlIjoiYzEifX19"`,
			body:            genericInvalidTokenBody,
			wantRetryable:   true,
		},
		{
			name:            "401 with interaction_required is retryable",
			status:          http.StatusUnauthorized,
			wwwAuthenticate: `Bearer error="interaction_required", error_description="additional auth required"`,
			body:            genericInvalidTokenBody,
			wantRetryable:   true,
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
