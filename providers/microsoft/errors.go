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
// ErrServer, etc.) with the Microsoft error-body parser above. handleErrorResponse
// delegates to this responder for every status except 401, where the extra
// inspection below is required.
var defaultJSONResponder = interpreter.NewFaultyResponder(errorFormats, nil) //nolint:gochecknoglobals

// handleErrorResponse classifies Microsoft error responses and is the single
// JSON error handler used by every Microsoft-family connector that shares
// this package (Microsoft / MicrosoftClientCredentials). It exists to fix a
// false-positive in the connection-status machinery upstream: the default
// interpreter maps every 401 to common.ErrAccessToken, and the server treats
// that as a signal to flip the connection to bad_credentials. Several
// legitimately-transient conditions at the Microsoft resource tier also
// surface as 401 — on those, flipping to bad_credentials is a user-visible
// regression that requires re-auth to clear.
//
// Classification strategy on 401:
//  1. If WWW-Authenticate signals a Continuous Access Evaluation (CAE) or
//     step-up challenge, the refresh token is still valid; a retry against
//     the freshly-re-issued token typically succeeds. Return ErrRetryable.
//  2. If the Graph error body matches a known transient marker (clock skew,
//     AAD replication race right after refresh), same verdict: ErrRetryable.
//  3. Otherwise treat as genuinely-invalid credentials and fall through to
//     the default responder, which will produce ErrAccessToken and — by
//     design — allow the server to flip the connection.
//
// Non-401 responses are unchanged from the prior behaviour and pass straight
// through to the default responder.
func handleErrorResponse(res *http.Response, body []byte) error {
	if res.StatusCode != http.StatusUnauthorized {
		return defaultJSONResponder.HandleErrorResponse(res, body)
	}

	if reason, ok := claimsChallengeReason(res.Header.Get("WWW-Authenticate")); ok {
		return fmt.Errorf("%w: microsoft claims challenge (%s)", common.ErrRetryable, reason)
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
// policy. In both cases the refresh token is still valid — the resource is
// asking for a new access token with extra claims, not signalling revoked
// credentials.
//
// Known limitation: a full CAE implementation would parse the claims="..."
// parameter out of the header and feed it back into the next token request
// so the re-issued access token satisfies the challenge. That requires
// plumbing through the token source and is out of scope here; we currently
// just mark the 401 as retryable and rely on Temporal retries to catch the
// eventual success once whatever transient condition clears. Sustained CAE
// challenges will still surface as activity failures after the retry budget
// is exhausted.
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
