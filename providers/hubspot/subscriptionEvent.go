package hubspot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type SubscriptionEvent struct {
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

// GetRecordFromSubscribeEvent fetches a record from the Hubspot API using the data from a subscription event.
func (c *Connector) GetRecordFromSubscriptionEvent(
	ctx context.Context, evt *SubscriptionEvent,
) (*common.ReadResultRow, error) {
	// Transform the subscription event into a ReadResult.
	objectName, err := evt.ObjectName()
	if err != nil {
		return nil, err
	}

	recordId := strconv.Itoa(evt.ObjectId)

	// Since the subscription event doesn't contain the record data, we need to fetch it.
	return c.GetRecord(ctx, objectName, recordId)
}

var errUnexpectedSubscriptionEventType = errors.New("unexpected subscription event type")

const minParts = 2

func (evt *SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	parts := strings.Split(evt.SubscriptionType, ".")

	if len(parts) < minParts {
		// this should never happen unless the provider changes subscription event format
		return common.SubscriptionEventTypeOther, fmt.Errorf(
			"%w: '%s'", errUnexpectedSubscriptionEventType, evt.SubscriptionType,
		)
	}

	switch parts[1] {
	case "creation":
		return common.SubscriptionEventTypeCreate, nil
	case "propertyChange":
		return common.SubscriptionEventTypeUpdate, nil
	case "deletion", "privacyDeletion":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt *SubscriptionEvent) RawEventName() (string, error) {
	return evt.SubscriptionType, nil
}

var errSubscriptionSupportedForObject = errors.New("subscription is not supported for the object")

func (evt *SubscriptionEvent) ObjectName() (string, error) {
	parts := strings.Split(evt.SubscriptionType, ".")
	if !getRecordSupportedObjectsSet.Has(parts[0]) {
		return "", fmt.Errorf("%w '%s'", errSubscriptionSupportedForObject, parts[0])
	}

	return parts[0], nil
}

func (evt *SubscriptionEvent) Workspace() (string, error) {
	return strconv.Itoa(evt.PortalId), nil
}

/*
	EXAMPLES: There is no documentation that shows data structure of subscription event.
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
