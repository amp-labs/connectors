package hubspot

import (
	"context"
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
	objectName := strings.Split(msg.SubscriptionType, ".")[0]
	recordId := strconv.Itoa(msg.ObjectId)

	// Since the webhook message doesn't contain the record data, we need to fetch it.
	return c.GetRecord(ctx, objectName, recordId)
}
