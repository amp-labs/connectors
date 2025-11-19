package outreach

import (
	"github.com/amp-labs/connectors/common"
)

type SubscriptionRequest struct {
	UniqueRef       string `json:"unique_ref"        validate:"required"`
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
	Secret          string `json:"secret,omitempty"`
}

type SubscriptionPayload struct {
	Data SubscriptionData `json:"data" validate:"required"`
}

type SubscriptionData struct {
	Type       string            `json:"type"       validate:"required"`
	Attributes AttributesPayload `json:"attributes" validate:"required"`
}

type AttributesPayload struct {
	Action   string `json:"action"   validate:"required"`
	Resource string `json:"resource" validate:"required"`
	URL      string `json:"url"      validate:"required"`
	Secret   string `json:"secret"`
}

type ModuleEvent string

var (
	Created   ModuleEvent = "created"   //nolint:gochecknoglobals
	Updated   ModuleEvent = "updated"   //nolint:gochecknoglobals
	Destroyed ModuleEvent = "destroyed" //nolint:gochecknoglobals
)

type createSubscriptionsResponse struct {
	Data createSubscriptionsResponseData `json:"data"`
}

type createSubscriptionsResponseData struct {
	// ID is the webhook subscription ID returned by Outreach API.
	// Outreach always returns this as a number (e.g., 15, 16, 17), not a string.
	ID         int            `json:"id"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

// SuccessfulSubscription is used internally for rollback tracking.
type SuccessfulSubscription struct {
	ID         string
	ObjectName string
	EventName  string
}

type SubscriptionResult struct {
	Subscriptions map[common.ObjectName]map[ModuleEvent]createSubscriptionsResponse `json:"Subscriptions"`
}
