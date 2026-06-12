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
	"github.com/amp-labs/connectors/providers"
)

var ErrSigningSecretIsNotSet = errors.New("signing secret is not set")

type Verifier struct {
	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	signingSecret string
}

// NewVerifier constructs an event message verifier.
// The empty signingSecret won't trigger a failure to preserve backward compatibility.
// However, no event message will be accepted.
func NewVerifier(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, signingSecret string) *Verifier {
	return &Verifier{
		client:        client,
		providerInfo:  providerInfo,
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
func (v Verifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	if v.signingSecret == "" {
		return false, ErrSigningSecretIsNotSet
	}

	slackSignature, err := httpkit.ExtractHeader(request.Headers, "X-Slack-Signature")
	if err != nil {
		return false, err
	}

	timestampStr, err := httpkit.ExtractHeader(request.Headers, "X-Slack-Request-Timestamp")
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
	h := hmac.New(sha256.New, []byte(v.signingSecret))
	h.Write([]byte(sigBasestring))
	computedSignature := "v0=" + hex.EncodeToString(h.Sum(nil))

	// Compare signatures using secure comparison
	return hmac.Equal([]byte(computedSignature), []byte(slackSignature)), nil
}

// abs returns the absolute value of an integer.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}

	return x
}
