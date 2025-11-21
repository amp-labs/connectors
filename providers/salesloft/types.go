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

type SubscriptionPayload struct {
	CallbackURL   string `json:"callback_url"        validate:"required"`
	EventType     string `json:"event_type" validate:"required"`
	CallbackToken string `json:"callback_token" validate:"required"`
}

type SubscriptionResponse struct {
	UserGuide     string `json:"user_guid"`
	TenantId      string `json:"tenant_id"`
	ID            int    `json:"id"`
	EventType     string `json:"event_type"`
	Enabled       bool   `json:"enabled"`
	CallbackURL   string `json:"callback_url"`
	CallbackToken string `json:"callback_token"`
}

type SuccessfulSubscription struct {
	ID         string
	ObjectName string
	EventName  string
}

type SubscriptionResult struct {
	Subscriptions map[common.ObjectName]map[ModuleEvent]SubscriptionResponse `json:"subscriptions"`
}

// ModuleEvent represents the combined event type format used by Salesloft.
// Format: "{objectName}_{eventAction}" (e.g., "person_created", "call_updated").
type ModuleEvent string

type EventAction string

const (
	ActionCreated EventAction = "created" //nolint:gochecknoglobals
	ActionUpdated EventAction = "updated" //nolint:gochecknoglobals
	ActionDeleted EventAction = "deleted" //nolint:gochecknoglobals
)

type SalesloftEventMapping struct {
	ObjectName      string
	SupportedEvents datautils.Set[ModuleEvent]
}
