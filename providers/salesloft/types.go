package salesloft

import (
	"github.com/amp-labs/connectors/common"
)

type subscriptionRequest struct {
	UniqueRef       string `json:"unique_ref"        validate:"required"`
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
	Secret          string `json:"secret,omitempty"`
}

type subscriptionPayload struct {
	CallbackURL   string `json:"callback_url"   validate:"required"`
	EventType     string `json:"event_type"     validate:"required"`
	CallbackToken string `json:"callback_token" validate:"required"`
}

type subscriptionResponse struct {
	UserGUID      string `json:"user_guid"`
	TenantId      string `json:"tenant_id"`
	ID            int    `json:"id"`
	EventType     string `json:"event_type"`
	Enabled       bool   `json:"enabled"`
	CallbackURL   string `json:"callback_url"`
	CallbackToken string `json:"callback_token"`
}
type successfulSubscription struct {
	ID         string
	ObjectName string
	EventName  string
}

type subscriptionResult struct {
	Subscriptions map[common.ObjectName]map[moduleEvent]subscriptionResponse `json:"subscriptions"`
}

// moduleEvent represents the combined event type string used by Salesloft.
// A moduleEvent value has the format "{objectName}_{eventAction}" (e.g., "person_created", "call_updated").
type moduleEvent string

type eventMapping struct {
	CreateEvents []moduleEvent
	UpdateEvents []moduleEvent
	DeleteEvents []moduleEvent
}

type salesloftObjectMapping struct {
	ObjectName string
	Events     eventMapping
}
