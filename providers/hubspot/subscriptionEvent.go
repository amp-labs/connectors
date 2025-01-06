package hubspot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

/*
Note:
SubscriptionEvent is a map[string]any as opposed to a typed struct, because the structure of the event is not known .
We may define the latest structure of the event, but in the future, the provider may add more fields.
In that case, we won't be receiving those fields in the event.
This form also prevents null fields to be sent out as zero values.
*/
type SubscriptionEvent map[string]any

// VerifyWebhookMessage verifies the signature of a webhook message from Hubspot.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context, params *common.WebhookVerificationParameters,
) (bool, error) {
	ts := params.Headers.Get(string(xHubspotRequestTimestamp))

	rawString := params.Method + params.URL + string(params.Body) + ts

	mac := hmac.New(sha256.New, []byte(params.ClientSecret))
	mac.Write([]byte(rawString))
	expectedMAC := mac.Sum(nil)

	signature := params.Headers.Get(string(xHubspotSignatureV3))

	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	return hmac.Equal(decodedSignature, expectedMAC), nil
}

var errUnexpectedSubscriptionEventType = errors.New("unexpected subscription event type")

const minParts = 2

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	subTypeStr, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	parts := strings.Split(subTypeStr, ".")

	if len(parts) < minParts {
		// this should never happen unless the provider changes subscription event format
		return common.SubscriptionEventTypeOther, fmt.Errorf(
			"%w: '%s'", errUnexpectedSubscriptionEventType, subTypeStr,
		)
	}

	switch parts[1] {
	case "creation":
		return common.SubscriptionEventTypeCreate, nil
	case "propertyChange":
		return common.SubscriptionEventTypeUpdate, nil
	case "deletion", "privacyDeletion":
		return common.SubscriptionEventTypeDelete, nil
	case "associationChange":
		return common.SubscriptionEventTypeUpdateAssociation, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	return m.GetString("subscriptionType")
}

var errSubscriptionSupportedForObject = errors.New("subscription is not supported for the object")

func (evt SubscriptionEvent) ObjectName() (string, error) {
	rawEvent, err := evt.RawEventName()
	if err != nil {
		return "", fmt.Errorf("error getting raw event name: %w", err)
	}

	parts := strings.Split(rawEvent, ".")
	if !getRecordSupportedObjectsSet.Has(parts[0]) {
		return "", fmt.Errorf("%w '%s'", errSubscriptionSupportedForObject, parts[0])
	}

	return parts[0], nil
}

func (evt SubscriptionEvent) Workspace() (string, error) {
	m := evt.asMap()

	portalId, err := m.AsInt("portalId")
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(portalId)), nil
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	m := evt.asMap()

	objId, err := m.AsInt("objectId")
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(objId)), nil
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	m := evt.asMap()

	ts, err := m.AsInt("occurredAt")
	if err != nil {
		return 0, err
	}

	return time.UnixMilli(ts).UnixNano(), nil
}

func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
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
