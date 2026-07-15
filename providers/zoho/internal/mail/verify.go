package mail

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// Zoho Mail uses the manual-config webhook pattern (SubscribeByAPI: false in the
// catalog): the outgoing webhook is created by hand in the Zoho Mail console
// (Settings > Integrations > Developer Space > Outgoing Webhooks), and this
// adapter only verifies deliveries (here), parses them into events
// (subscriptionEvent.go), and fetches full records on demand (records.go). Zoho
// Mail has no API for the programmatic Subscribe/Update/Delete lifecycle (unlike
// Zoho CRM), so those operations do not exist here.
//
// Docs: https://www.zoho.com/mail/help/dev-platform/webhook.html

// mailHookSignatureHeader is the per-request signature header Zoho Mail attaches
// to every webhook delivery. Its value is
// base64(HMAC-SHA256(x-hook-secret, rawRequestBody)).
const mailHookSignatureHeader = "X-Hook-Signature"

var ErrMissingWebhookSecret = errors.New("zoho mail webhook secret is not set")

// VerifyWebhookMessage validates that a webhook request came from Zoho Mail by
// recomputing the HMAC-SHA256 signature over the raw request body with the
// signing secret and comparing it (constant-time) to the X-Hook-Signature
// header.
func (a *Adapter) VerifyWebhookMessage(
	_ context.Context, request *common.WebhookRequest, _ *common.VerificationParams,
) (bool, error) {
	if a.hookSecret == "" {
		return false, ErrMissingWebhookSecret
	}

	got, err := httpkit.ExtractRequiredHeader(request.Headers, mailHookSignatureHeader)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(a.hookSecret))
	mac.Write(request.Body)
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(want), []byte(got)), nil
}
