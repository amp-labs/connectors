package salesloft

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

type SubscriptionRequest struct {
	UniqueRef       string `json:"unique_ref"         validate:"required"`
	WebhookEndPoint string `json:"webhook_end_point"  validate:"required"`
	Secret          string `json:"secret,omitempty"`
}

// SubscriptionPayload represents the payload sent to Salesloft's webhook API
type SubscriptionPayload struct {
	CallbackURL   string `json:"callback_url"        validate:"required"`
	EventType     string `json:"event_type" validate:"required"`
	CallbackToken string `json:"callback_token" validate:"required"`
}

// SubscriptionResponse represents the response from Salesloft's webhook API
type SubscriptionResponse struct {
	UserGuide     string `json:"user_guid"`
	TenantId      string `json:"tenant_id"`
	ID            int    `json:"id"`
	EventType     string `json:"event_type"`
	Enabled       bool   `json:"enabled"`
	CallbackURL   string `json:"callback_url"`
	CallbackToken string `json:"callback_token"`
}

// SuccessfulSubscription tracks successful subscriptions for rollback purposes
type SuccessfulSubscription struct {
	ID         string
	ObjectName string
	EventName  string
}

// SubscriptionResult stores the final subscription results
type SubscriptionResult struct {
	Subscriptions map[common.ObjectName]map[SalesloftEventType]SubscriptionResponse `json:"subscriptions"`
}

// SalesloftEventType represents the combined event type format used by Salesloft
// Format: "{objectName}_{eventAction}" (e.g., "person_created", "call_updated")
type SalesloftEventType string

// Base event actions used in Salesloft
type EventAction string

const (
	ActionCreated EventAction = "created" //nolint:gochecknoglobals
	ActionUpdated EventAction = "updated" //nolint:gochecknoglobals
	ActionDeleted EventAction = "deleted" //nolint:gochecknoglobals
)

type SalesloftEventMapping struct {
	ObjectName      string                            // singular form used by Salesloft
	SupportedEvents datautils.Set[SalesloftEventType] // actual Salesloft event names supported (O(1) lookup)
}
