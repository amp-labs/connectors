package attio

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type (
	SubscriptionEvent map[string]any
	//nolint: godoclint
	// Attio sends Secret in response when we subscribe to webhooks.
	// We use this secret to verify the webhook signatures.
	AttioVerificationParams struct {
		Secret string
	}
)

var (
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
)

const (
	signatureHeader = "attio-signature"
)

// CollapsedSubscriptionEvent represents the raw webhook payload from attio.
// Unlike Salesforce or Zoho, attio sends individual events (one record per webhook),
// so this implementation simply wraps the single event.
type CollapsedSubscriptionEvent map[string]any

// RawMap returns a copy of the raw event data.
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// During testing we observed that Attion sends only one event but wraps it in an array.
// So We are extracting that single event and returning it as a list.
// if in future Attio changes this behavior to send multiple events in one payload,
// this code will still work.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {

	subscribeEvents := make([]common.SubscriptionEvent, 0)

	m := common.StringMap(e)

	events, err := m.Get("events")
	if err != nil {
		return nil, err
	}

	eventsArr, ok := events.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected []any, got %T", errTypeMismatch, eventsArr)
	}

	for index, evt := range eventsArr {
		evtMap, ok := evt.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: expected map[string]any at index %d, got %T", errTypeMismatch, index, evt)
		}

		subscribeEvents = append(subscribeEvents, SubscriptionEvent(evtMap))
	}
	return subscribeEvents, nil
}

func (evt SubscriptionEvent) PreLoadData(data *common.SubscriptionEventPreLoadData) error {
	return nil
}

// VerifyWebhookMessage implements WebhookVerifierConnector for Attio.
// Returns (true, nil) if signature verification succeeds.
// Returns (false, error) if verification fails or encounters an error.
// Ref: https://docs.attio.com/rest-api/guides/webhooks#authenticating
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

	signature := request.Headers.Get(signatureHeader)
	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, signatureHeader)
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

func computeSignature(secret string, body []byte) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)

	return h.Sum(nil)
}

func (s SubscriptionEvent) EventTimeStampNano() (int64, error) {

	// Attio does not provide event timestamp in webhook response.
	// So we are returning zero value.
	return 0, nil
}

// ObjectName implements [common.SubscriptionUpdateEvent].
func (s SubscriptionEvent) ObjectName() (string, error) {

	return "", nil
}

// EventType implements [common.SubscriptionUpdateEvent].
func (s SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	m := s.asMap()

	eventType, err := m.GetString("event_type")
	if err != nil {
		return "", err
	}

	return toCommonEvent(providerEvent(eventType))
}

// RawEventName implements [common.SubscriptionUpdateEvent].
func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	eventType, err := m.GetString("event_type")
	if err != nil {
		return "", err
	}

	return eventType, nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	rawEventName, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	m := evt.asMap()

	idField, err := m.Get("id")
	if err != nil {
		return "", fmt.Errorf("failed to get id field :%v", err)
	}

	idMap, ok := idField.(map[string]string)
	if !ok {
		return "", fmt.Errorf("%w:%s expected map[string]string, got %T", errTypeMismatch, "IdMap", idMap)
	}

	idKey := strings.Split(rawEventName, ".")[0] + "_id"

	recordId, ok := idMap[idKey]
	if err != nil {
		return "", fmt.Errorf("failed to get record id :%v", err)
	}

	return recordId, nil
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	// Attio does not provide updated fields in webhook response.
	return []string{}, nil
}

// Workspace is not available in Attio.
func (evt SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
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
