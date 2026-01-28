package attio

type subscriptionRequest struct {
	WebhookEndpoint string `json:"webhook_end_point" validate:"required"`
}

// Reference: https://docs.attio.com/rest-api/endpoint-reference/webhooks/create-a-webhook#body-data
type subscriptionPayload struct {
	Data subscriptionData `json:"data" validate:"required"`
}

type subscriptionData struct {
	TargetURL     string         `json:"target_url"    validate:"required"`
	Subscriptions []subscription `json:"subscriptions" validate:"required"`
}

type subscription struct {
	EventType providerEvent `json:"event_type" validate:"required"`
	// Filter is an object used to limit which webhook events are delivered.
	// Filters can target specific records (by list_id, entry_id) and specific
	// It cannot be used to do field level filtering.
	// Use null to receive all events without filtering.
	// Ref: https://docs.attio.com/rest-api/guides/webhooks#filtering
	Filter any `json:"filter"     validate:"required"`
}

// Reference: https://docs.attio.com/rest-api/endpoint-reference/webhooks/create-a-webhook#response-data
type createSubscriptionsResponse struct {
	Data createSubscriptionsResponseData `json:"data"`
}

type createSubscriptionsResponseID struct {
	WorkspaceID string `json:"workspace_id"`
	WebhookID   string `json:"webhook_id"`
}

type createSubscriptionsResponseData struct {
	TargetURL     string                        `json:"target_url"`
	Subscriptions []subscription                `json:"subscriptions" validate:"required"`
	ID            createSubscriptionsResponseID `json:"id"`
	Status        string                        `json:"status"`
	CreatedAt     string                        `json:"created_at"`
}

// SuccessfulSubscription is used internally for rollback tracking.
type SuccessfulSubscription struct {
	ID         string
	ObjectName string
	EventName  string
}

type subscriptionResult struct {
	Data createSubscriptionsResponseData `json:"data"`
}

// providerEvent represents the combined event type string used by Attio.
// A providerEvent value has the format "{objectName}.{eventAction}" (e.g., note.created", "task.updated").
type providerEvent string

type objectEvents struct {
	createEvents []providerEvent
	updateEvents []providerEvent
	deleteEvents []providerEvent
}
