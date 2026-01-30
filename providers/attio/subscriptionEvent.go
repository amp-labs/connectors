package attio

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type (
	SubscriptionEvent map[string]any
	//nolint: godoclint
	// Attio sends Secret in response when we subscribe to webhooks.
	// We use this secret to verify the webhook signatures.
	AttioVerificationParams struct {
		Secret string
	}
)

const (
	signatureHeader = "attio-signature"
)

// VerifyWebhookMessage implements WebhookVerifierConnector for Attio.
// Returns (true, nil) if signature verification succeeds.
// Returns (false, error) if verification fails or encounters an error.
// Ref: https://docs.attio.com/rest-api/guides/webhooks#authenticating
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*AttioVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	signature := request.Headers.Get(signatureHeader)
	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, signatureHeader)
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("%w: error decoding signature: %w", ErrInvalidSignature, err)
	}

	expectedSignature := computeSignature(verificationParams.Secret, request.Body)

	if !hmac.Equal(sigBytes, expectedSignature) {
		return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
	}

	return true, nil
}

func computeSignature(secret string, body []byte) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)

	return h.Sum(nil)
}

// Example: Webhook response
/*
{
 "webhook_id": "04731154-70d3-42bb-8320-760304c9bbfd",
 "events": [
   {
     "event_type": "note.updated",
     "id": {
       "workspace_id": "e293215c-210a-4d4a-9913-e2b33da318ab",
       "note_id": "f83d5cab-571b-47a8-8018-57146f848d19"
     },
     "parent_object_id": "ee1e6aa1-ec69-4ef4-a101-3a9abb12e281",
     "parent_record_id": "9bcad14b-55a5-478d-963b-a4ec598265c6",
     "actor": {
       "type": "workspace-member",
       "id": "f0519378-80b8-4d7c-8874-c6acc1850442"
     }
   }
 ]
}
*/
