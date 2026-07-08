package jobber

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

// Jobber webhook payloads are minimal: they carry only the topic, account and
// item identifiers, so record data must be fetched back via GetRecordsByIds.
//
// Example payload:
//
//	{
//	  "data": {
//	    "webHookEvent": {
//	      "topic": "CLIENT_CREATE",
//	      "appId": "3ef22a50-072d-430c-a78f-b7646657560b",
//	      "accountId": "MQ==",
//	      "itemId": "MQ==",
//	      "occurredAt": "2021-08-12T16:31:36-06:00"
//	    }
//	  }
//	}
//
// Reference: https://developer.getjobber.com/docs/using_jobbers_api/setting_up_webhooks/

const (
	// WebhookSignatureHeader carries Base64(HMAC-SHA256(client secret, raw body)).
	// Jobber sends "X-Jobber-Hmac-SHA256" on the wire; the constant uses Go's
	// canonical form, which http.Header lookups normalize to either way.
	WebhookSignatureHeader = "X-Jobber-Hmac-Sha256"
)

var (
	ErrMissingSignature = errors.New("missing webhook signature header")
	ErrInvalidSignature = errors.New("invalid webhook signature")

	errMissingVerificationParams = errors.New("jobber: missing verification params")
	errEventTypeMismatch         = errors.New("jobber: unexpected type in webhook event field")
	errMissingEventField         = errors.New("jobber: missing webhook event field")
)

type (
	// SubscriptionEvent wraps a single parsed Jobber webhook payload.
	SubscriptionEvent map[string]any

	// CollapsedSubscriptionEvent represents the raw webhook payload. Jobber
	// sends one event per webhook request, so this simply wraps the single event.
	CollapsedSubscriptionEvent map[string]any

	// JobberVerificationParams carries the secret used to verify webhook
	// signatures. Jobber signs payloads with the app's OAuth client secret.
	JobberVerificationParams struct {
		Secret string `json:"secret" validate:"required"`
	}
)

var (
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
)

// VerifyWebhookMessage checks the X-Jobber-Hmac-SHA256 header:
// Base64(HMAC-SHA256(key = app client secret, message = raw request body)).
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingVerificationParams)
	}

	verificationParams, err := common.AssertType[*JobberVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingVerificationParams, err)
	}

	signature := request.Headers.Get(WebhookSignatureHeader)
	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, WebhookSignatureHeader)
	}

	mac := hmac.New(sha256.New, []byte(verificationParams.Secret))
	mac.Write(request.Body)

	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
	}

	return true, nil
}

// RawMap returns a copy of the raw event data.
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList returns the event as a single-element list.
// Jobber webhooks contain exactly one event per payload.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

func (evt SubscriptionEvent) PreLoadData(*common.SubscriptionEventPreLoadData) error {
	return nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

// RawEventName returns the Jobber topic, e.g. "CLIENT_CREATE".
func (evt SubscriptionEvent) RawEventName() (string, error) {
	return evt.eventField("topic")
}

// EventType derives the CRUD event type from the topic suffix. Topics
// outside the CREATE/UPDATE/DESTROY convention (QUOTE_SENT, JOB_CLOSED,
// VISIT_COMPLETE, ...) are reported as "other".
func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	topic, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, err
	}

	switch {
	case strings.HasSuffix(topic, "_CREATE"):
		return common.SubscriptionEventTypeCreate, nil
	case strings.HasSuffix(topic, "_UPDATE"):
		return common.SubscriptionEventTypeUpdate, nil
	case strings.HasSuffix(topic, "_DESTROY"):
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

// ObjectName maps the topic back to the connector object name, e.g.
// CLIENT_CREATE -> clients, QUOTE_SENT -> quotes. Topics that have no
// corresponding connector object (APP_CONNECT, PAYMENT_*, ...) return the
// raw topic so callers can still identify the event source.
func (evt SubscriptionEvent) ObjectName() (string, error) {
	topic, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	for obj, root := range objectTopicRoot {
		if strings.HasPrefix(topic, root+"_") && validTopics.Has(topic) {
			return obj.String(), nil
		}
	}

	return topic, nil
}

// Workspace returns the Jobber account ID the event originated from.
// A single app receives events from every account that installed it, so the
// account ID is the tenant discriminator.
func (evt SubscriptionEvent) Workspace() (string, error) {
	return evt.eventField("accountId")
}

// RecordId returns the encoded ID of the affected record; the full record
// must be fetched back via the GraphQL API.
func (evt SubscriptionEvent) RecordId() (string, error) {
	return evt.eventField("itemId")
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	// Apps created before 2023-12-08 receive the misspelled "occuredAt" key.
	timestamp, err := evt.eventField("occurredAt")
	if err != nil {
		timestamp, err = evt.eventField("occuredAt")
		if err != nil {
			return 0, err
		}
	}

	occurredAt, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0, fmt.Errorf("error parsing occurredAt time: %w", err)
	}

	return occurredAt.UnixNano(), nil
}

// eventField extracts a string field from the data.webHookEvent object.
func (evt SubscriptionEvent) eventField(field string) (string, error) {
	data, isMap := evt["data"].(map[string]any)
	if !isMap {
		return "", fmt.Errorf("%w: data", errMissingEventField)
	}

	event, isMap := data["webHookEvent"].(map[string]any)
	if !isMap {
		return "", fmt.Errorf("%w: data.webHookEvent", errMissingEventField)
	}

	value, exists := event[field]
	if !exists {
		return "", fmt.Errorf("%w: data.webHookEvent.%s", errMissingEventField, field)
	}

	str, isString := value.(string)
	if !isString {
		return "", fmt.Errorf("%w: data.webHookEvent.%s expected string, got %T",
			errEventTypeMismatch, field, value)
	}

	return str, nil
}
