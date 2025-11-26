package attio

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type (
	SubscriptionEvent map[string]any

	// Attio sends Secret in response when we subscribe to webhooks.
	// We use this secret to verify the webhook signatures.
	AttioVerificationParams struct {
		Secret string
	}
)

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}

	errTypeMismatch = errors.New("type mismatch")
)

const (
	SignatureHeader = "attio-signature"
)

// VerifyWebhookMessage implements WebhookVerifierConnector for Attio.
// Returns (true, nil) if signature verification succeeds.
// Returns (false, error) if verification fails or encounters an error.
// Note: Return type changed from error to (bool, error) to match the interface contract.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*AttioVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	signature := request.Headers.Get(SignatureHeader)
	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, SignatureHeader)
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("%w: error decoding signature: %w", ErrInvalidSignature, err)
	}

	expectedSignature := computeSignature(verificationParams.Secret, request.Body)

	if !hmac.Equal(sigBytes, expectedSignature) {
		return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
	}

	return true, nil
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	return nil, errors.New("attio webhooks do not provide updated field information") //nolint:err113
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	return 0, errors.New("attio webhooks do not include event timestamps") //nolint:err113
}

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	subTypeStr, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	parts := strings.Split(subTypeStr, ".")
	if len(parts) < 2 { //nolint:mnd
		// this should never happen unless the provider changes subscription event format
		return common.SubscriptionEventTypeOther, fmt.Errorf(
			"%w: '%s'", errUnsupportedEventType, subTypeStr,
		)
	}

	switch parts[1] {
	case string(Created):
		return common.SubscriptionEventTypeCreate, nil
	case string(Updated):
		return common.SubscriptionEventTypeUpdate, nil
	case string(Deleted):
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	m := evt.asMap()

	events, err := m.Get("events")
	if err != nil {
		return "", err
	}

	eventsMap, ok := events.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, eventsMap, events)
	}

	name, ok := eventsMap["event_type"].(string)
	if !ok {
		return "", fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, name, eventsMap["event_type"])
	}

	objectName := strings.Split(name, ".")[0]

	return objectName, nil
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	events, err := m.Get("events")
	if err != nil {
		return "", err
	}

	eventsMap, ok := events.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, eventsMap, events)
	}

	eventName, ok := eventsMap["event_type"].(string)
	if !ok {
		return "", fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, eventName, eventsMap["event_type"])
	}

	return eventName, nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	m := evt.asMap()

	events, err := m.Get("events")
	if err != nil {
		return "", err
	}

	eventsMap, exist := events.(map[string]any)
	if !exist {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, eventsMap, events)
	}

	idMap, exist := eventsMap["id"].(map[string]string)
	if !exist {
		return "", fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, idMap, eventsMap["id"])
	}

	objectName, err := evt.ObjectName()
	if err != nil {
		return "", err
	}

	idKey := objectName + "_id"

	id, ok := idMap[idKey]
	if !ok {
		return "", fmt.Errorf("idMap does not contain id %s", idKey) //nolint:err113
	}

	return id, nil
}

// No workspace concept in Outreach.
func (evt SubscriptionEvent) Workspace() (string, error) {
	m := evt.asMap()

	data, err := m.Get("events")
	if err != nil {
		return "", err
	}

	dataMap, exist := data.(map[string]any)
	if !exist {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, dataMap, data) // nolint:err113
	}

	idMap, ok := dataMap["id"].(map[string]string)
	if !ok {
		return "", fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, idMap, dataMap["id"]) // nolint:err113
	}

	id, ok := idMap["workspace_id"]
	if !ok {
		return "", fmt.Errorf("idMap does not contain id %s", "workspace_id") //nolint:err113
	}

	return id, nil
}

// asMap returns the event as a StringMap.
func (evt SubscriptionEvent) asMap() common.StringMap {
	// extract first event from events array
	// Attio sends an array of events, but it only contains one event per webhook call.
	// So we extract the first event for processing.
	evtsArry, ok := evt["events"].([]any)
	if ok && len(evtsArry) > 0 {
		firstEvt, ok := evtsArry[0].(map[string]any)
		if ok {
			return common.StringMap(firstEvt)
		}
	}

	// Fallback to returning the whole event if extraction fails
	return common.StringMap(evt)
}

func computeSignature(secret string, body []byte) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)

	return h.Sum(nil)
}

// Example: Webhook response
/*
{
  "webhook_id": "04731154-70d3-42bb-8320-760304c9bbfd",
  "events": [
    {
      "event_type": "note.updated",
      "id": {
        "workspace_id": "e293215c-210a-4d4a-9913-e2b33da318ab",
        "note_id": "f83d5cab-571b-47a8-8018-57146f848d19"
      },
      "parent_object_id": "ee1e6aa1-ec69-4ef4-a101-3a9abb12e281",
      "parent_record_id": "9bcad14b-55a5-478d-963b-a4ec598265c6",
      "actor": {
        "type": "workspace-member",
        "id": "f0519378-80b8-4d7c-8874-c6acc1850442"
      }
    }
  ]
}
*/
