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

type WebhookResult struct {
	WebhookMessage *WebhookMessage    `json:"webhookMessage"`
	Record         *common.ReadResult `json:"record"`
}

func (c *Connector) GetWebhookResultFromWebhookMessage(
	ctx context.Context, msg *WebhookMessage,
) (*WebhookResult, error) {
	// Transform the webhook message into a ReadResult.
	objectName := strings.Split(msg.SubscriptionType, ".")[0]
	recordId := strconv.Itoa(msg.ObjectId)

	// Since the webhook message doesn't contain the record data, we need to fetch it.
	recordResult, err := c.GetRecord(ctx, objectName, recordId)
	if err != nil {
		return nil, err
	}

	return &WebhookResult{
		WebhookMessage: msg,
		Record:         recordResult,
	}, nil
}
