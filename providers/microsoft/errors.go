package microsoft

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// errorFormats registers the single canonical Microsoft error envelope with
// the interpreter framework. Graph, Outlook, and similar APIs all use this
// shape; nil MustKeys means "always match.".
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

// ResponseMessageError is the canonical Microsoft Graph / REST error shape.
// See https://learn.microsoft.com/en-us/graph/errors.
type ResponseMessageError struct {
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		InnerError any    `json:"innerError"`
	} `json:"error"`
}

// CombineErr wraps the status-code sentinel with the provider-side message so
// errors.Is(err, ErrAccessToken) still works while the log carries context.
func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Error.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Error.Message)
}

// defaultJSONResponder applies the standard status-code → sentinel mapping.
// handleErrorResponse delegates to it for everything except the two special
// 401 cases below.
var defaultJSONResponder = interpreter.NewFaultyResponder(errorFormats, nil) //nolint:gochecknoglobals

// handleErrorResponse classifies Microsoft 401s to avoid flipping connections
// to bad_credentials on conditions that aren't actually bad credentials.
//
// Rule: only mark a 401 retryable when a retry with the same refresh token
// could plausibly succeed. Anything needing user action falls through to the
// bad-credentials path.
//
//  1. Transient marker in the error body  → common.ErrRetryable.
//  2. CAE / step-up claim challenge       → ClaimsChallengeError
//     (wraps common.ErrAccessToken; server still flips bad_credentials).
//  3. Everything else                     → default responder (ErrAccessToken).
//
// Non-401 responses are unchanged.
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

// claimsChallengeReason detects CAE (insufficient_claims) and step-up
// (interaction_required) challenges in WWW-Authenticate. These are not
// retryable — a plain refresh can't satisfy a claim challenge; the user has
// to re-authenticate. Detection exists to attach diagnostic context.
//
// See https://learn.microsoft.com/en-us/entra/identity-platform/claims-challenge
// for the challenge format, and the ClaimsChallengeError type for how the
// result is propagated.
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

// transientAuthErrorMessage recognises Graph 401 bodies that indicate a
// short-lived failure (clock skew between token-issuing and resource regions,
// or AAD JWKS replication lag right after a refresh) rather than a truly
// invalid token. A retry seconds later almost always succeeds.
//
// Gated on Error.Code == "InvalidAuthenticationToken" plus one of the
// observed transient messages. Intentionally conservative — excluded patterns
// like "CompactToken validation failed" also appear for real failures, and
// misclassifying them as retryable would mask legitimate bad credentials.
// Extend `markers` only after production logs confirm a phrase correlates
// exclusively with transient cases.
//
// See https://learn.microsoft.com/en-us/graph/errors for the envelope.
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

// ClaimsChallengeError is returned for Microsoft CAE and step-up challenges.
// It wraps common.ErrAccessToken (via Is) so the server's existing
// bad-credentials flow still fires — semantically correct, the user has to
// re-authenticate. The struct is exported so the server can errors.As for a
// future distinct re-auth status; for now, callers get structured detail via
// the Error() string without needing to know about this type.
//
//   - Reason: "insufficient_claims" (CAE) or "interaction_required" (step-up).
//   - WWWAuthenticate: raw header; contains a base64 claims="..." payload
//     operators can decode to see what Azure is demanding.
//   - RequestMethod / RequestURL: which Graph endpoint rejected the token.
//   - MSRequestID: x-ms-request-id, for correlation with Azure sign-in logs.
type ClaimsChallengeError struct {
	Reason          string
	WWWAuthenticate string
	RequestMethod   string
	RequestURL      string
	MSRequestID     string
}

// Error packs every field into one string so the existing server-side log
// pattern `logger.Error("...", "error", err)` surfaces all diagnostic detail
// without needing an errors.As block on the server. Format:
//
//	microsoft claims challenge (<reason>) [<METHOD> <URL>] [x-ms-request-id=<id>] [www-authenticate=<header>]
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

// Is matches common.ErrAccessToken so the server's existing auth-invalidated
// path handles claim challenges. Deliberately does NOT match ErrRetryable.
func (e *ClaimsChallengeError) Is(target error) bool {
	return target == common.ErrAccessToken
}

// newClaimsChallengeError tolerates a nil res.Request (synthesised responses
// in tests) by leaving method/url empty rather than panicking.
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
