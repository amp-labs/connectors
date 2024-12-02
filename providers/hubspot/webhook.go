package hubspot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type WebhookMessage struct {
	AppId            int    `json:"appId"`
	EventId          int    `json:"eventId"`
	SubscriptionId   int    `json:"subscriptionId"`
	PortalId         int    `json:"portalId"`
	OccurredAt       int    `json:"occurredAt"`
	SubscriptionType string `json:"subscriptionType"`
	AttemptNumber    int    `json:"attemptNumber"`
	ObjectId         int    `json:"objectId"`
	ChangeSource     string `json:"changeSource"`
	PropertyName     string `json:"propertyName"`
	PropertyValue    string `json:"propertyValue"`
}

// GetRecordFromWebhookMessage fetches a record from the Hubspot API using the data from a webhook message.
func (c *Connector) GetRecordFromWebhookMessage(
	ctx context.Context, msg *WebhookMessage,
) (*common.ReadResultRow, error) {
	// Transform the webhook message into a ReadResult.
	objectName, err := msg.ObjectName()
	if err != nil {
		return nil, err
	}

	recordId := strconv.Itoa(msg.ObjectId)

	// Since the webhook message doesn't contain the record data, we need to fetch it.
	return c.GetRecord(ctx, objectName, recordId)
}

var errUnexpectedWebhookEventType = errors.New("unexpected webhook event type")

const minParts = 2

func (msg *WebhookMessage) EventType() (common.WebhookEventType, error) {
	parts := strings.Split(msg.SubscriptionType, ".")

	if len(parts) < minParts {
		// this should never happen unless the provider changes webhook message format
		return common.WebhookEventTypeOther, fmt.Errorf("%w: '%s'", errUnexpectedWebhookEventType, msg.SubscriptionType)
	}

	switch parts[1] {
	case "creation":
		return common.WebhookEventTypeCreate, nil
	case "propertyChange":
		return common.WebhookEventTypeUpdate, nil
	case "deletion", "privacyDeletion":
		return common.WebhookEventTypeDelete, nil
	default:
		return common.WebhookEventTypeOther, nil
	}
}

func (msg *WebhookMessage) RawEventName() (string, error) {
	return msg.SubscriptionType, nil
}

var errWebhookNotSupportedForObject = errors.New("webhook is not supported for the object")

func (msg *WebhookMessage) ObjectName() (string, error) {
	parts := strings.Split(msg.SubscriptionType, ".")
	if !getRecordSupportedObjectsSet.Has(parts[0]) {
		return "", fmt.Errorf("%w '%s'", errWebhookNotSupportedForObject, parts[0])
	}

	return parts[0], nil
}

func (msg *WebhookMessage) Workspace() (string, error) {
	return strconv.Itoa(msg.PortalId), nil
}

/*
	EXAMPLES: There is no documentation that shows data structure of webhook messages.
	Below examples were found from hubspot app settings page after login at:
	https://app.hubspot.com/private-apps/<<CustomerAppId>>/<<PrivateAppId>>/webhooks
	Or from UI on Customer Account
	Settings -> Account Management -> Integrations -> Private Apps -> <<YOUR PRIVATE APP>> -> Webhooks

	{
		"appId": 4210286,
		"eventId": 100,
		"subscriptionId": 2881778,
		"portalId": 44237313,
		"occurredAt": 1731612159499,
		"subscriptionType": "contact.creation",
		"attemptNumber": 0,
		"objectId": 123,
		"changeSource": "CRM",
		"changeFlag": "NEW"
	}

	{
		"appId": 4210286,
		"eventId": 100,
		"subscriptionId": 2902227,
		"portalId": 44237313,
		"occurredAt": 1731612210994,
		"subscriptionType": "contact.propertyChange",
		"attemptNumber": 0,
		"objectId": 123,
		"changeSource": "CRM",
		"propertyName": "message",
		"propertyValue": "sample-value"
	}
*/
