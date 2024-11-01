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
	OccurredAt       string `json:"occurredAt"`
	SubscriptionType string `json:"subscriptionType"`
	AttemptNumber    int    `json:"attemptNumber"`
	ObjectId         int    `json:"objectId"`
	ChangeSource     string `json:"changeSource"`
	PropertyName     string `json:"propertyName"`
	PropertyValue    string `json:"propertyValue"`
}

func (c *Connector) TransformWebhookMessageToReadResult(ctx context.Context, msg WebhookMessage) (*common.ReadResult, error) {
	// Transform the webhook message into a ReadResult.
	// This is a placeholder implementation.
	objectName := strings.Split(msg.SubscriptionType, ".")[0]
	recordId := strconv.Itoa(msg.ObjectId)

	recordResult, err := c.GetRecord(ctx, objectName, recordId)
	if err != nil {
		return nil, err
	}

	return recordResult, nil
}
