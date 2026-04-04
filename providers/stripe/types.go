package stripe

import (
	"time"

	"github.com/amp-labs/connectors/common"
)

// SubscriptionRequest represents the parameters required to subscribe a Stripe webhook
// endpoint via this connector's public API. The WebhookEndPoint field must be the full
// HTTPS URL that will receive Stripe webhook events.
//
// For more details on Stripe webhook endpoints, see:
// https://docs.stripe.com/api/webhook_endpoints
type SubscriptionRequest struct {
	// WebhookEndPoint is the full HTTPS URL where Stripe will send webhook events.
	// See: https://docs.stripe.com/api/webhook_endpoints/create
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
}

// WebhookPayload is the payload sent to Stripe when creating or updating
// a webhook endpoint.
//
// For more details, see:
// https://docs.stripe.com/api/webhook_endpoints/create
//
// Example payload:
//
//	{
//	  "url": "https://example.com/api/stripe/webhook",
//	  "enabled_events": ["customer.created", "customer.updated"]
//	}
type WebhookPayload struct {
	// URL is the endpoint's URL. See: https://docs.stripe.com/api/webhook_endpoints/create
	URL string `json:"url" validate:"required"`
	// EnabledEvents are the events to enable for this endpoint.
	// See: https://docs.stripe.com/api/webhook_endpoints/create#create_webhook_endpoint-enabled_events
	// For available event types, see: https://docs.stripe.com/api/events/types
	EnabledEvents []string `json:"enabled_events" validate:"required"`
}

// WebhookResponse is the response from Stripe when creating, retrieving,
// or updating a webhook endpoint.
//
// For more details on the webhook endpoint object, see:
// https://docs.stripe.com/api/webhook_endpoints/object
//
// Example response:
//
//	{
//		"id": "we_1Mr5jULkdIwHu7ix1ibLTM0x",
//		"object": "webhook_endpoint",
//		"api_version": null,
//		"application": null,
//		"created": 1680122196,
//		"description": null,
//		"enabled_events": [
//		  "charge.succeeded",
//		  "charge.failed"
//		],
//		"livemode": false,
//		"metadata": {},
//		"secret": "whsec_wRNftLajMZNeslQOP6vEPm4iVx5NlZ6z",
//		"status": "enabled",
//		"url": "https://example.com/my/webhook/endpoint"
//	  }
type WebhookResponse struct {
	ID            string   `json:"id"`
	Object        string   `json:"object"`
	URL           string   `json:"url"`
	EnabledEvents []string `json:"enabled_events"`
	Status        string   `json:"status"`
	Secret        string   `json:"secret,omitempty"`
}

// SubscriptionResult contains the result of a subscription operation, mapping
// object names to their webhook endpoint responses.
type SubscriptionResult struct {
	Subscriptions map[common.ObjectName]WebhookResponse `json:"subscriptions"`
}

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
