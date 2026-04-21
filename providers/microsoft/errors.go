package microsoft

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

type ResponseMessageError struct {
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		InnerError any    `json:"innerError"`
	} `json:"error"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Error.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Error.Message)
}

// defaultJSONResponder handles non-401 responses with the standard Microsoft
// error-body parsing and default status-code mapping.
var defaultJSONResponder = interpreter.NewFaultyResponder(errorFormats, nil) //nolint:gochecknoglobals

// handleErrorResponse classifies Microsoft Graph error responses. On 401 it
// distinguishes transient conditions (CAE claim challenges, token lifetime
// races right after a refresh, AAD replication lag) from truly-invalid
// credentials, returning common.ErrRetryable for the former so that the
// server's workflow layer retries rather than flipping the connection to
// bad_credentials. All other status codes fall through to the default
// responder.
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
// error="insufficient_claims" for CAE and error="interaction_required" when
// the resource requires user interaction to satisfy a policy. The refresh
// token itself is still valid in these cases; the resource is asking for a
// new token with extra claims, not signalling revoked credentials.
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
// These typically occur immediately after a successful refresh due to clock
// skew between the token-issuing region and the resource region, or due to
// AAD signing-key replication lag after key rotation. A retry with the same
// (or a freshly re-minted) token usually succeeds.
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
