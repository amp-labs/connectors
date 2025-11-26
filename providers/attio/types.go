package attio

type SubscriptionRequest struct {
	UniqueRef       string `json:"unique_ref"        validate:"required"`
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
	Secret          string `json:"secret,omitempty"`
}

type SubscriptionPayload struct {
	Data SubscriptionData `json:"data" validate:"required"`
}

type SubscriptionData struct {
	TargetURL     string         `json:"target_url"       validate:"required"`
	Subscriptions []Subscription `json:"subscriptions" validate:"required"`
}

type Subscription struct {
	EventType string `json:"event_type"   validate:"required"`
}

type ModuleEvent string

var (
	Created ModuleEvent = "created" //nolint:gochecknoglobals
	Updated ModuleEvent = "updated" //nolint:gochecknoglobals
	Deleted ModuleEvent = "deleted" //nolint:gochecknoglobals
)

type createSubscriptionsResponse struct {
	Data createSubscriptionsResponseData `json:"data"`
}

type createSubscriptionsResponseId struct {
	WorkspaceID string `json:"workspace_id"`
	WebhookID   string `json:"webhook_id"`
}

type createSubscriptionsResponseData struct {
	TargetURL     string                        `json:"target_url"`
	Subscriptions []Subscription                `json:"subscriptions" validate:"required"`
	ID            createSubscriptionsResponseId `json:"id"`
	Status        string                        `json:"status"`
	CreatedAt     string                        `json:"created_at"`
}

// SuccessfulSubscription is used internally for rollback tracking.
type SuccessfulSubscription struct {
	ID         string
	ObjectName string
	EventName  string
}

type SubscriptionResult struct {
	Data createSubscriptionsResponseData `json:"Data"`
}
