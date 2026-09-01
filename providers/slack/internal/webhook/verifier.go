package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

var ErrSigningSecretIsNotSet = errors.New("signing secret is not set")

// VerificationParams carries the caller-resolved inputs for Slack webhook verification.
// SigningSecret is the Slack app's signing secret (Basic Information > App Credentials in the
// Slack app dashboard); in the Ampersand subscribe flow it is collected from the builder by the
// dashboard and stored in ProviderApp.metadata.providerParams.
type VerificationParams struct {
	SigningSecret string
}

type Verifier struct {
	signingSecret string
}

// NewVerifier constructs an event message verifier.
// The empty signingSecret won't trigger a failure to preserve backward compatibility.
// However, no event message will be accepted.
func NewVerifier(signingSecret string) *Verifier {
	return &Verifier{
		signingSecret: signingSecret,
	}
}

// VerifyWebhookMessage validates that the webhook request came from Slack by verifying
// the HMAC-SHA256 signature using the signing secret.
//
// Reference: https://docs.slack.dev/authentication/verifying-requests-from-slack/
//
// Slack signs requests with:
//   - X-Slack-Signature header: v0=<hex-digest>
//   - X-Slack-Request-Timestamp header: Unix timestamp
//
// The signature is computed as:
//
//	sig_basestring = "v0:" + timestamp + ":" + request_body
//	signature = "v0=" + HMAC-SHA256(signing_secret, sig_basestring).hex()
//
// The request is rejected if the timestamp is more than 5 minutes old (replay attack protection).
//
// The signing secret is resolved from params (VerificationParams.SigningSecret) when provided,
// falling back to the secret the Verifier was constructed with. The params path is what the
// subscribe flow uses: its verifier connector is a zero value (no construction metadata), so the
// secret must arrive per-request — hence the pointer receiver and nil-safety.
func (v *Verifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	secret, err := v.resolveSigningSecret(params)
	if err != nil {
		return false, err
	}

	slackSignature, err := httpkit.ExtractRequiredHeader(request.Headers, "X-Slack-Signature")
	if err != nil {
		return false, err
	}

	timestampStr, err := httpkit.ExtractRequiredHeader(request.Headers, "X-Slack-Request-Timestamp")
	if err != nil {
		return false, err
	}

	// Validate timestamp is not more than 5 minutes old (replay attack protection)
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp format: %w", err)
	}

	if abs(time.Now().Unix()-timestamp) > 5*60 {
		return false, errors.New("request timestamp is more than 5 minutes old") // nolint:err113
	}

	// Build the signature basestring: v0:timestamp:request_body
	requestBody := string(request.Body)
	sigBasestring := fmt.Sprintf("v0:%s:%s", timestampStr, requestBody)

	// Compute HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(sigBasestring))
	computedSignature := "v0=" + hex.EncodeToString(h.Sum(nil))

	// Compare signatures using secure comparison
	return hmac.Equal([]byte(computedSignature), []byte(slackSignature)), nil
}

// resolveSigningSecret picks the signing secret for a verification call: a non-empty secret in
// params wins, then the constructor-provided secret. Nil-safe on the receiver so the zero-value
// slack.Connector used by the subscribe seam (whose embedded *Verifier is nil) still verifies
// with params-provided secrets.
func (v *Verifier) resolveSigningSecret(params *common.VerificationParams) (string, error) {
	if params != nil && params.Param != nil {
		verificationParams, err := common.AssertType[*VerificationParams](params.Param)
		if err != nil {
			return "", fmt.Errorf("invalid verification params: %w", err)
		}

		if verificationParams.SigningSecret != "" {
			return verificationParams.SigningSecret, nil
		}
	}

	if v != nil && v.signingSecret != "" {
		return v.signingSecret, nil
	}

	return "", ErrSigningSecretIsNotSet
}

// abs returns the absolute value of an integer.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}

	return x
}
