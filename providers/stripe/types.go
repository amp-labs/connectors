package stripe

import (
	"time"
)

// VerificationParams contains the parameters needed to verify Stripe webhook signatures.
//
// For more details on webhook signature verification, see:
// https://docs.stripe.com/webhooks
type VerificationParams struct {
	// Secret is the webhook signing secret used to verify the HMAC signature of incoming
	// webhook requests. This secret is provided by Stripe when creating a webhook endpoint.
	// See: https://docs.stripe.com/webhooks?verify=verify-manually
	Secret string

	// Tolerance is the maximum allowed time difference between the webhook timestamp
	// and the current time. This prevents replay attacks by rejecting old webhook payloads.
	// Defaults to 5 minutes if not specified. Must be greater than 0.
	// Stripe generates a new timestamp and signature for each delivery attempt, including retries.
	// See: https://docs.stripe.com/webhooks/signatures#verify-manually
	Tolerance time.Duration
}
