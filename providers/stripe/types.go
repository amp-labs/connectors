package stripe

import (
	"github.com/amp-labs/connectors/common"
)

// SubscriptionRequest represents the parameters required to subscribe a Stripe webhook.
type SubscriptionRequest struct {
	// WebhookEndPoint is the full HTTPS URL where Stripe will send webhook events.
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
}

// WebhookEndpointPayload is the payload sent to stripe.
type WebhookEndpointPayload struct {
	URL           string   `json:"url"            validate:"required"`
	EnabledEvents []string `json:"enabled_events" validate:"required"`
}

// WebhookEndpointResponse is the response from stripe.
type WebhookEndpointResponse struct {
	ID            string   `json:"id"`
	Object        string   `json:"object"`
	URL           string   `json:"url"`
	EnabledEvents []string `json:"enabled_events"`
	Status        string   `json:"status"`
	Secret        string   `json:"secret,omitempty"`
}

type SubscriptionResult struct {
	Subscriptions map[common.ObjectName]WebhookEndpointResponse `json:"subscriptions"`
}

// StripeVerificationParams contains the parameters needed to verify Stripe webhook signatures.
type StripeVerificationParams struct {
	// Secret is the webhook signing secret used to verify the HMAC signature of incoming
	// webhook requests. This secret is provided by Stripe when creating a webhook endpoint.
	Secret string
}
