package stripe

import (
	"github.com/amp-labs/connectors/common"
)

type SubscriptionRequest struct {
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
}

// WebhookEndpointPayload is the payload sent to stripe.
type WebhookEndpointPayload struct {
	URL     string   `json:"url"            validate:"required"`
	Enabled []string `json:"enabled_events" validate:"required"`
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
	Subscriptions map[common.ObjectName]WebhookEndpointResponse `json:"Subscriptions"`
}

// StripeVerificationParams contains the secret needed to verify Stripe webhook signatures.
type StripeVerificationParams struct {
	Secret string
}
