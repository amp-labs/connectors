package outreach

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

type (
	SubscriptionEvent          map[string]any
	OutreachVerificationParams struct {
		Secret string
	}
)

var (
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}

	errTypeMismatch = errors.New("type mismatch")
)

// CollapsedSubscriptionEvent represents the raw webhook payload from Outreach.
// Unlike Salesforce or Zoho, Outreach sends individual events (one record per webhook),
// so this implementation simply wraps the single event.
type CollapsedSubscriptionEvent map[string]any

// RawMap returns a copy of the raw event data.
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList returns the event as a single-element list.
// Outreach webhooks contain only one record per payload, so no fan-out is needed.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

const (
	OutreachWebhookSignatureHeader = "Outreach-Webhook-Signature"
)

// VerifyWebhookMessage implements WebhookVerifierConnector for Outreach.
// Returns (true, nil) if signature verification succeeds.
// Returns (false, error) if verification fails or encounters an error.
// Note: Return type changed from error to (bool, error) to match the interface contract.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	// Verify the webhook message
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*OutreachVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	signature := request.Headers.Get(OutreachWebhookSignatureHeader)
	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, OutreachWebhookSignatureHeader)
	}

	expectedSignature := computeSignature(verificationParams.Secret, request.Body)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
	}

	return true, nil
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	m := evt.asMap()

	data, err := m.Get("data")
	if err != nil {
		return nil, err
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected %T got %T", errTypeMismatch, dataMap, data)
	}

	attributes, ok := dataMap["attributes"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected %T got %T", errTypeMismatch, attributes, dataMap["attributes"])
	}

	updatedFields := make([]string, 0, len(attributes))

	for field := range attributes {
		updatedFields = append(updatedFields, field)
	}

	return updatedFields, nil
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	m := evt.asMap()

	meta, err := m.Get("meta")
	if err != nil {
		return 0, err
	}

	metaMap, ok := meta.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("%w: expected %T got %T", errTypeMismatch, metaMap, meta)
	}

	deliveredAtStr, ok := metaMap["deliveredAt"].(string)
	if !ok {
		return 0, fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, deliveredAtStr, metaMap["deliveredAt"])
	}

	deliveredAt, err := time.Parse(time.RFC3339Nano, deliveredAtStr)
	if err != nil {
		return 0, fmt.Errorf("error parsing deliveredAt time: %w", err)
	}

	return deliveredAt.UnixNano(), nil
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
			"%w: '%s'", errUnexpectedSubscriptionEventType, subTypeStr,
		)
	}

	switch parts[1] {
	case "created":
		return common.SubscriptionEventTypeCreate, nil
	case "updated":
		return common.SubscriptionEventTypeUpdate, nil
	case "destroyed":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	m := evt.asMap()

	data, err := m.Get("data")
	if err != nil {
		return "", err
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, dataMap, data)
	}

	name, ok := dataMap["type"].(string)
	if !ok {
		return "", fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, name, dataMap["type"])
	}

	return name, nil
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	meta, err := m.Get("meta")
	if err != nil {
		return "", err
	}

	metaMap, ok := meta.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, metaMap, meta)
	}

	eventName, ok := metaMap["eventName"].(string)
	if !ok {
		return "", fmt.Errorf("%w: expected %T, got %T", errTypeMismatch, eventName, metaMap["eventName"])
	}

	return eventName, nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	m := evt.asMap()

	data, err := m.Get("data")
	if err != nil {
		return "", err
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: expected %T got %T", errTypeMismatch, dataMap, data)
	}

	// Outreach sends numeric IDs in webhook payloads.
	// JSON unmarshals numbers as float64, so we handle both string and numeric types.
	switch id := dataMap["id"].(type) {
	case string:
		return id, nil
	case float64:
		return fmt.Sprintf("%.0f", id), nil
	default:
		return "", fmt.Errorf("%w: expected string or number, got %T", errTypeMismatch, dataMap["id"])
	}
}

// Workspace returns an empty string as there is no workspace concept in Outreach.
func (evt SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

// asMap returns the event as a StringMap.
func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
}

func computeSignature(secret string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)

	return hex.EncodeToString(h.Sum(nil))
}

// Example: Webhook response
/*
{
  "data": {
    "type": "account",
    "id": 13,
    "attributes": {
      "createdAt": "2025-11-04T09:40:36.000Z",
      "updatedAt": "2025-11-04T09:40:36.000Z",
      "named": true,
      "domain": "test.com",
      "externalSource": "outreach-api",
      "name": "this is a test"
    },
    "relationships": {
      "owner": {
        "type": "owner",
        "id": 2
      }
    }
  },
  "meta": {
    "deliveredAt": "2025-11-04T01:40:36.795-08:00",
    "eventName": "account.created",
    "jobId": "13dc9ab5-5ccc-4fbb-bdf9-cdbcdd986621"
  }
}
*/
