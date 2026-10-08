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

type VerificationParams struct {
	SigningSecret string
}

type Verifier struct{}

// errVerifierNotInitialized is returned when VerifyWebhookMessage is called through a Connector
// whose embedded *Verifier was never constructed.
var errVerifierNotInitialized = errors.New("slack webhook verifier is not initialized")

// NewVerifier constructs an event message verifier.
// The empty signingSecret won't trigger a failure to preserve backward compatibility.
// However, no event message will be accepted.
func NewVerifier() *Verifier {
	return &Verifier{}
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
// Pointer receiver with an explicit nil guard: a slack.Connector whose embedded *Verifier was
// never initialized (e.g. a zero-value literal) must fail verification with a clear error, never
// panic in the method-promotion wrapper. The subscribe registry constructs the Verifier via
// slack.NewWebhookVerifierConnector, so the guard is a backstop, not the expected path.
func (v *Verifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	if v == nil {
		return false, errVerifierNotInitialized
	}

	signingSecret, err := v.resolveSigningSecret(params)
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
	h := hmac.New(sha256.New, []byte(signingSecret))
	h.Write([]byte(sigBasestring))
	computedSignature := "v0=" + hex.EncodeToString(h.Sum(nil))

	// Compare signatures using secure comparison
	return hmac.Equal([]byte(computedSignature), []byte(slackSignature)), nil
}

func (v *Verifier) resolveSigningSecret(params *common.VerificationParams) (string, error) {
	if params == nil || params.Param == nil {
		return "", common.ErrMissingVerificationParams
	}

	verificationParams, err := common.AssertType[*VerificationParams](params.Param)
	if err != nil {
		return "", fmt.Errorf("%w: %w", common.ErrInvalidVerificationParams, err)
	}

	if verificationParams.SigningSecret == "" {
		return "", fmt.Errorf("%w: SigningSecret is empty", common.ErrMissingProviderParam)
	}

	return verificationParams.SigningSecret, nil
}

// abs returns the absolute value of an integer.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}

	return x
}
