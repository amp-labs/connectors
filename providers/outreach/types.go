package outreach

type SubscriptionRequest struct {
	UniqueRef       string `json:"unique_ref"        validate:"required"`
	WebhookEndPoint string `json:"webhook_end_point" validate:"required"`
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
	ID         string                        `json:"id"`
	Type       string                        `json:"type"`
	Attributes createSubscriptionsAttributes `json:"attributes"`
}

type createSubscriptionsAttributes struct {
	Action         string `json:"action"`
	Active         bool   `json:"active"`
	CleanupToken   string `json:"cleanupToken"`
	URL            string `json:"url"`
	CreatedAt      string `json:"createdAt"`
	Resource       string `json:"resource"`
	Secret         string `json:"secret"`
	PayloadVersion string `json:"payloadVersion"`
	CreatorAppName string `json:"creatorAppName"`
	CreatorAppId   string `json:"creatorAppId"`
	DisabledReason string `json:"disabledReason"`
	DisabledSince  string `json:"disabledSince"`
	DisabledUntil  string `json:"disabledUntil"`
}

type SuccessfulSubscription struct {
	ID         string `json:"id"`
	ObjectName string `json:"object_name"`
	EventName  string `json:"event_name"`
}

type FailedSubscription struct {
	ObjectName string `json:"object_name"`
	EventName  string `json:"event_name"`
	Error      string `json:"error"` // Use string instead of error for JSON serialization
}

type SubscriptionResultData struct {
	SuccessfulSubscriptions []SuccessfulSubscription `json:"successful_subscriptions"`
	FailedSubscriptions     []FailedSubscription     `json:"failed_subscriptions"`
}
