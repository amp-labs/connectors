package microsoft

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// errorFormats is the JSON schema set used by the interpreter framework to
// parse Microsoft error payloads. Microsoft Graph, Outlook, and similar
// services wrap errors in a single canonical envelope (ResponseMessageError),
// so we register exactly one template. FormatSwitch picks the best-matching
// template by MustKeys; nil keys mean "always match this template" when no
// other candidates are registered.
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

// ResponseMessageError is the canonical error envelope used by Microsoft
// Graph and sibling Microsoft REST APIs. See:
// https://learn.microsoft.com/en-us/graph/errors
//
// The Code field is a machine-readable classifier (e.g. "InvalidAuthenticationToken",
// "Forbidden", "ResourceNotFound"); Message is a human-readable description whose
// wording can vary across regions and over time — see transientAuthErrorMessage
// for notes on matching against it. InnerError carries additional diagnostic
// detail we currently only surface via generic unmarshalling.
type ResponseMessageError struct {
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		InnerError any    `json:"innerError"`
	} `json:"error"`
}

// CombineErr satisfies the interpreter.ErrorDescriptor contract. The
// framework calls it with the status-code-derived sentinel (e.g.
// common.ErrBadRequest, common.ErrAccessToken) and expects us to enrich it
// with provider-specific context. We wrap the base sentinel with the Graph
// message so that downstream callers see, e.g., "access token invalid: Lifetime
// validation failed..." while still being able to errors.Is(err, ErrAccessToken).
func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Error.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Error.Message)
}

// defaultJSONResponder handles any response whose error classification is
// unambiguous from the HTTP status code alone. It combines the standard
// status-code mapping (401 -> ErrAccessToken, 403 -> ErrForbidden, 5xx ->
// ErrServer, etc.) with the Microsoft error-body parser above.
// handleErrorResponse delegates to this responder for all non-401 responses,
// and for 401s that don't match either the transient-marker case or the
// claim-challenge case — i.e. anything that really is an invalid token.
var defaultJSONResponder = interpreter.NewFaultyResponder(errorFormats, nil) //nolint:gochecknoglobals

// handleErrorResponse classifies Microsoft error responses and is the single
// JSON error handler used by every Microsoft-family connector that shares
// this package (Microsoft / MicrosoftClientCredentials). It exists to fix a
// false-positive in the connection-status machinery upstream: the default
// interpreter maps every 401 to common.ErrAccessToken, and the server treats
// that as a signal to flip the connection to bad_credentials. Some 401s
// actually reflect short-lived, self-healing conditions at the Microsoft
// resource tier — on those, flipping to bad_credentials is a user-visible
// regression that requires re-auth to clear.
//
// Guiding rule: a 401 is only marked retryable if a retry with the same (or
// a freshly re-minted) refresh token has a realistic chance of succeeding
// without any external change. Anything that demands user action or
// code-level change must *not* be retryable, because retries would just
// burn the budget and delay surfacing the real problem.
//
// Classification strategy on 401:
//  1. If the Graph error body matches a known transient marker (clock skew,
//     AAD JWKS replication race right after refresh), the next attempt
//     almost always succeeds — return common.ErrRetryable.
//  2. If WWW-Authenticate signals a Continuous Access Evaluation (CAE) or
//     step-up claim challenge, the refresh token is still usable but the
//     challenge itself demands new claims that a plain refresh cannot
//     obtain (admin session revoke, password change, MFA step-up,
//     compliant-device requirement, etc.). Return a ClaimsChallengeError
//     that wraps common.ErrAccessToken via the Is method — so the server
//     flips the connection to bad_credentials (semantically correct: the
//     user must re-authenticate) while still exposing the structured
//     details for logging and future UI differentiation.
//  3. Otherwise fall through to the default responder, which produces
//     common.ErrAccessToken — by design, the server flips the connection.
//
// Non-401 responses are unchanged from the prior behaviour and pass straight
// through to the default responder.
func handleErrorResponse(res *http.Response, body []byte) error {
	if res.StatusCode != http.StatusUnauthorized {
		return defaultJSONResponder.HandleErrorResponse(res, body)
	}

	if reason, ok := claimsChallengeReason(res.Header.Get("WWW-Authenticate")); ok {
		return newClaimsChallengeError(res, reason)
	}

	if msg, ok := transientAuthErrorMessage(body); ok {
		return fmt.Errorf("%w: transient microsoft auth error: %s", common.ErrRetryable, msg)
	}

	return defaultJSONResponder.HandleErrorResponse(res, body)
}

// claimsChallengeReason detects a Continuous Access Evaluation (CAE) or
// step-up claim challenge in the WWW-Authenticate header. Microsoft uses
// error="insufficient_claims" for CAE (e.g. session revoked mid-flight by
// admin action, risk signal, password change) and error="interaction_required"
// when the resource requires user interaction to satisfy a Conditional Access
// policy.
//
// These challenges are not transient — the refresh token by itself can't
// satisfy the requested claims. MFA step-up needs the user; compliant-device
// checks need the user's device; admin session revocation needs the user to
// re-authenticate to create a fresh session. We therefore classify claim
// challenges as non-retryable and fall through to the bad-credentials path
// (see ClaimsChallengeError). The detection here exists to attach structured
// diagnosis details, not to enable retries.
//
// References:
//   - Claims challenges / claims requests / CAE client handling:
//     https://learn.microsoft.com/en-us/entra/identity-platform/claims-challenge
//   - Continuous Access Evaluation (overview + resilience guidance):
//     https://learn.microsoft.com/en-us/entra/identity-platform/app-resilience-continuous-access-evaluation
//   - CAE in Conditional Access (operator-facing explanation of which events
//     cause insufficient_claims):
//     https://learn.microsoft.com/en-us/entra/identity/conditional-access/concept-continuous-access-evaluation
//   - interaction_required is the OIDC standard "step-up" response, used by
//     Entra when a Conditional Access policy demands MFA/compliant device/etc.:
//     https://openid.net/specs/openid-connect-core-1_0.html#AuthError
//
// Future work: a full CAE implementation would parse the claims="..."
// parameter out of the header and feed it back into the next token request
// so the re-issued access token satisfies the challenge. That would move
// some (not all) of these challenges from "needs user re-auth" to "the
// system can auto-satisfy" — e.g., service-principal ACR requirements.
// Even then, challenges requiring real user interaction (MFA, compliant
// device) remain non-retryable by definition.
//
// The matching is deliberately case-insensitive and uses substring lookup
// rather than full RFC 6750 parsing because the Microsoft header format is
// stable enough and we only care about the two known values.
func claimsChallengeReason(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	lower := strings.ToLower(header)
	switch {
	case strings.Contains(lower, `error="insufficient_claims"`):
		return "insufficient_claims", true
	case strings.Contains(lower, `error="interaction_required"`):
		return "interaction_required", true
	}

	return "", false
}

// transientAuthErrorMessage recognises Graph error bodies that indicate a
// transient authentication failure rather than a permanently-invalid token.
// These typically occur immediately after a successful refresh due to:
//   - Clock skew between the token-issuing region (login.microsoftonline.com
//     shard that minted the access token) and the resource region (Graph
//     shard validating it) — Azure's ~5-minute `nbf`/`exp` allowance is
//     occasionally exceeded under load.
//   - Azure AD signing-key replication lag after a key rotation, where the
//     resource has not yet fetched the new JWKS entry.
//
// In both cases a retry with the same (or a freshly re-minted) token usually
// succeeds within seconds.
//
// References:
//   - Microsoft Graph error response schema (documents the `error.code` /
//     `error.message` envelope matched here):
//     https://learn.microsoft.com/en-us/graph/errors
//   - InvalidAuthenticationToken in the Graph error code reference:
//     https://learn.microsoft.com/en-us/graph/errors#code-property
//   - AAD signing key rollover (explains the JWKS replication race that
//     produces transient signature/lifetime validation failures on freshly
//     minted tokens):
//     https://learn.microsoft.com/en-us/entra/identity-platform/signing-key-rollover
//   - Entra ID (AAD) error code reference — companion to Graph errors, lists
//     the AADSTS codes that can accompany InvalidAuthenticationToken:
//     https://learn.microsoft.com/en-us/entra/identity-platform/reference-error-codes
//   - JWT `nbf` / `exp` semantics that underlie the "not yet valid" and
//     "lifetime validation failed" messages (RFC 7519 §4.1.4–4.1.5):
//     https://www.rfc-editor.org/rfc/rfc7519#section-4.1.4
//
// The matcher is intentionally conservative:
//   - We require Error.Code == "InvalidAuthenticationToken"; a message that
//     looks similar under a different code is more likely a real auth issue.
//   - The message markers are drawn from observed transient error text. We
//     deliberately exclude broader patterns like "CompactToken validation
//     failed" because those are also emitted for genuinely-malformed tokens
//     and would risk masking real failures.
//   - Matching is substring-based, not exact — Microsoft varies the trailing
//     text (e.g. "... the token is expired" vs "... token is expired and can
//     no longer be used") across regions and versions.
//
// Extend `markers` only after confirming from production logs that a given
// phrase exclusively correlates with a transient condition; misclassifying a
// real auth failure as retryable delays user-visible error reporting.
func transientAuthErrorMessage(body []byte) (string, bool) {
	var messageError ResponseMessageError
	if err := json.Unmarshal(body, &messageError); err != nil {
		return "", false
	}

	if !strings.EqualFold(messageError.Error.Code, "InvalidAuthenticationToken") {
		return "", false
	}

	markers := []string{
		"Lifetime validation failed",
		"Access token has expired",
		"The token is not yet valid",
	}

	for _, marker := range markers {
		if strings.Contains(messageError.Error.Message, marker) {
			return messageError.Error.Message, true
		}
	}

	return "", false
}

// ClaimsChallengeError is returned by handleErrorResponse when Microsoft
// responds with a Continuous Access Evaluation or step-up claim challenge
// on a 401.
//
// Claim challenges are NOT retryable. They indicate the user's session is
// still usable but the resource is demanding additional claims (re-auth,
// MFA step-up, compliant-device etc.) that cannot be obtained by simply
// repeating the request with the same refresh token. Retrying would just
// burn the retry budget and delay the eventual re-auth that the user has
// to perform. Accordingly this type wraps common.ErrAccessToken via the Is
// method, so the server flips the connection to bad_credentials via its
// existing auth-invalidated path — the user-visible signal "your connection
// needs to be re-established" is correct for a claim challenge even if the
// underlying cause is subtler than a revoked refresh token.
//
// The type exists so the server can still identify these events
// specifically — via errors.As — and log or surface them differently from
// a plain bad-credentials flip. That gives us the hook for a future
// CONNECTION_STATUS_REAUTH_REQUIRED (or similarly-differentiated state)
// without having to change the classification here.
//
// Example server-side use:
//
//	chErr, ok := errors.AsType[*microsoft.ClaimsChallengeError](err)
//	if ok {
//	    logger.Warn("microsoft claims challenge",
//	        "reason", chErr.Reason,
//	        "url", chErr.RequestURL,
//	        "x_ms_request_id", chErr.MSRequestID,
//	        "www_authenticate", chErr.WWWAuthenticate)
//	}
//
// Field notes:
//   - Reason: which of the two WWW-Authenticate error codes fired —
//     "insufficient_claims" (CAE) or "interaction_required" (step-up).
//   - WWWAuthenticate: the raw header. The claims="..." parameter inside is
//     a base64url-encoded JSON object describing exactly what claims Azure
//     is demanding; operators can decode it to understand the challenge.
//   - RequestMethod / RequestURL: identify which Graph endpoint rejected
//     the token. Useful when a single connection drives multiple objects.
//   - MSRequestID: Microsoft's x-ms-request-id. Pair with Azure sign-in
//     logs to find the Conditional Access policy that produced the
//     challenge.
type ClaimsChallengeError struct {
	Reason          string
	WWWAuthenticate string
	RequestMethod   string
	RequestURL      string
	MSRequestID     string
}

// Error serialises every diagnostic field into a single string so that
// existing server-side error logs — e.g. shared/workflow/read/error.go
// which logs `"error", err` on the ErrAccessToken branch — surface all
// the details automatically, without the server needing to perform an
// errors.As and re-log individual fields. The format is:
//
//	microsoft claims challenge (<reason>) [<METHOD> <URL>] [x-ms-request-id=<id>] [www-authenticate=<header>]
//
// Bracketed sections are omitted when the underlying field is empty.
func (e *ClaimsChallengeError) Error() string {
	if e == nil {
		return ""
	}

	var challengeBuilder strings.Builder

	fmt.Fprintf(&challengeBuilder, "microsoft claims challenge (%s)", e.Reason)

	if e.RequestURL != "" {
		fmt.Fprintf(&challengeBuilder, " [%s %s]", e.RequestMethod, e.RequestURL)
	}

	if e.MSRequestID != "" {
		fmt.Fprintf(&challengeBuilder, " [x-ms-request-id=%s]", e.MSRequestID)
	}

	if e.WWWAuthenticate != "" {
		fmt.Fprintf(&challengeBuilder, " [www-authenticate=%s]", e.WWWAuthenticate)
	}

	return challengeBuilder.String()
}

// Is lets errors.Is(err, common.ErrAccessToken) return true for this type,
// so the server's existing auth-invalidated / bad-credentials detection
// continues to fire for claim challenges without needing to know about
// this concrete type. Claim challenges explicitly do *not* satisfy
// errors.Is(err, common.ErrRetryable) — see the type-level comment for the
// reasoning.
func (e *ClaimsChallengeError) Is(target error) bool {
	return target == common.ErrAccessToken
}

// newClaimsChallengeError extracts the pieces of the response that are
// useful for diagnosis and packs them into a ClaimsChallengeError. The
// response's Request may be nil if the response was synthesized (e.g.,
// in tests), in which case method/url are left empty rather than panicking.
func newClaimsChallengeError(res *http.Response, reason string) *ClaimsChallengeError {
	err := &ClaimsChallengeError{
		Reason:          reason,
		WWWAuthenticate: res.Header.Get("WWW-Authenticate"),
		MSRequestID:     res.Header.Get("X-Ms-Request-Id"),
	}

	if res.Request != nil {
		err.RequestMethod = res.Request.Method
		if res.Request.URL != nil {
			err.RequestURL = res.Request.URL.String()
		}
	}

	return err
}
