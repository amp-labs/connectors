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
type SubscriptionEvent struct {
	AppId            int    `json:"appId"`
	EventId          int    `json:"eventId"`
	SubscriptionId   int    `json:"subscriptionId"`
	PortalId         int    `json:"portalId"`
	OccurredAt       int    `json:"occurredAt"` // in milliseconds
	SubscriptionType string `json:"subscriptionType"`
	AttemptNumber    int    `json:"attemptNumber"`
	ChangeSource     string `json:"changeSource"`
	// Optional fields
	ObjectId   *int    `json:"objectId,omitempty"`
	ChangeFlag *string `json:"changeFlag,omitempty"`
	// Property Change Fields
	PropertyName  *string `json:"propertyName,omitempty"`
	PropertyValue *string `json:"propertyValue,omitempty"`
	// Association Change Fields
	AssociationType      *string `json:"associationType,omitempty"`
	FromObjectId         *int    `json:"fromObjectId,omitempty"`
	ToObjectId           *int    `json:"toObjectId,omitempty"`
	AssociationRemoved   *bool   `json:"associationRemoved,omitempty"`
	IsPrimaryAssociation *bool   `json:"isPrimaryAssociation,omitempty"`
}


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

var (
	errUnexpectedSubscriptionEventType = errors.New("unexpected subscription event type")
	errSubscriptionTypeNotFound        = errors.New("subscription type not found")
	errFieldTypeMismatch               = errors.New("field type mismatch")
)

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
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	subType, ok := evt["subscriptionType"]
	if !ok {
		return "", errSubscriptionTypeNotFound
	}

	subTypeStr, ok := subType.(string)
	if !ok {
		return "", fmt.Errorf(
			"%w: expecting string but got '%T'", errFieldTypeMismatch, subType,
		)
	}

	return subTypeStr, nil
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

var errNotFound = errors.New("property name not found")

func (evt SubscriptionEvent) Workspace() (string, error) {
	portalId, ok := evt["portalId"]
	if !ok {
		return "", fmt.Errorf("%w: portalId", errNotFound)
	}

	portalIdInt, ok := portalId.(float64)
	if !ok {
		return "", fmt.Errorf("portalId %w, expected int, but received '%T'", errFieldTypeMismatch, portalId)
	}

	idInt := int(portalIdInt)

	return strconv.Itoa(idInt), nil
}

var errRecordIdNotAvailable = errors.New("record ID is not available")

func (evt SubscriptionEvent) RecordId() (string, error) {
	objIdRaw, ok := evt["objectId"]
	if !ok {
		return "", errRecordIdNotAvailable
	}

	objId, ok := objIdRaw.(float64)
	if !ok {
		return "", fmt.Errorf("objectId %w, expected int, but received '%T'", errFieldTypeMismatch, objIdRaw)
	}

	objIdInt := int(objId)

	return strconv.Itoa(objIdInt), nil
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	tsRaw, ok := evt["occurredAt"]
	if !ok {
		return 0, fmt.Errorf("%w: occurredAt", errNotFound)
	}

	ts, ok := tsRaw.(float64)
	if !ok {
		return 0, fmt.Errorf("occurredAt %w, expected int, but received '%T'", errFieldTypeMismatch, ts)
	}

	tsInt := int64(ts)

	return time.UnixMilli(tsInt).UnixNano(), nil
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
