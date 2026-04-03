package housecallpro

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"maps"
	"regexp"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	housecallAPITimestampHeader = "Api-Timestamp"
	housecallAPISignatureHeader = "Api-Signature"
)

var (
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
	// eventTypeRegex validates event types with at least 2 dot-separated parts (e.g. job.created).
	eventTypeRegex = regexp.MustCompile(`^[^.]+(\.[^.]+)+$`)
)

// HousecallProVerificationParams configures webhook HMAC verification (see Housecall webhooks docs).
type HousecallProVerificationParams struct{ Secret string }

// SubscriptionEvent represents a webhook event from Housecall Pro.
type SubscriptionEvent map[string]any

// CollapsedSubscriptionEvent is the raw webhook payload. Housecall sends one event per request.
type CollapsedSubscriptionEvent map[string]any

//nolint:gochecknoglobals // Static allowlist of webhook objects we support.
var supportedEventObjects = datautils.NewSetFromList([]string{"jobs", "customers", "estimates", "invoices", "leads"})

// VerifyWebhookMessage implements connectors.WebhookVerifierConnector.
func (*Connector) VerifyWebhookMessage(
	_ context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*HousecallProVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	timestamp := request.Headers.Get(housecallAPITimestampHeader)
	if timestamp == "" {
		return false, fmt.Errorf("%w: missing %s header", errMissingParams, housecallAPITimestampHeader)
	}

	sig := request.Headers.Get(housecallAPISignatureHeader)
	if sig == "" {
		return false, fmt.Errorf("%w: missing %s header", errMissingParams, housecallAPISignatureHeader)
	}

	mac := hmac.New(sha256.New, []byte(verificationParams.Secret))
	mac.Write([]byte(timestamp + "." + string(request.Body)))

	if housecallSignaturesEqual(sig, mac.Sum(nil)) {
		return true, nil
	}

	return false, fmt.Errorf("%w", errInvalidSignature)
}

func housecallSignaturesEqual(headerSignature string, expected []byte) bool {
	headerSignature = strings.TrimSpace(headerSignature)
	headerSignature = strings.TrimPrefix(strings.ToLower(headerSignature), "sha256=")

	b, err := hex.DecodeString(headerSignature)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(b, expected) == 1
}

func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	eventType, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	if !eventTypeRegex.MatchString(eventType) {
		return common.SubscriptionEventTypeOther, nil
	}

	parts := strings.Split(eventType, ".")

	action := strings.ToLower(parts[len(parts)-1])
	switch action {
	case "created":
		return common.SubscriptionEventTypeCreate, nil
	case "updated":
		return common.SubscriptionEventTypeUpdate, nil
	case "deleted", "destroyed":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	return evt.asMap().GetString("event")
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	name, _, err := resourceRecord(evt)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (evt SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	raw, err := evt.asMap().GetString("event_occurred_at")
	if err != nil {
		return 0, err
	}

	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return 0, fmt.Errorf("parse event_occurred_at: %w", err)
	}

	return t.UnixNano(), nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	node, err := jsonquery.Convertor.NodeFromMap(evt)
	if err != nil {
		return "", fmt.Errorf("failed to convert event to node: %w", err)
	}

	resource, err := evt.eventResourcePrefix()
	if err != nil {
		return "", err
	}

	payloadObject, _, err := resolveEventObject(resource)
	if err != nil {
		return "", err
	}

	return jsonquery.New(node, payloadObject).StringRequired("id")
}

// UpdatedFields implements common.SubscriptionUpdateEvent.
// Housecall Pro does not indicate which fields changed in webhook payloads.
func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	return []string{}, nil
}

func (evt SubscriptionEvent) PreLoadData(*common.SubscriptionEventPreLoadData) error {
	return nil
}

func (evt SubscriptionEvent) asMap() common.StringMap { return common.StringMap(evt) }

func (evt SubscriptionEvent) splitEventType() (resource string, err error) {
	raw, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	if !eventTypeRegex.MatchString(raw) {
		return "", fmt.Errorf("%w: %q", errInvalidEventTypeFormat, raw)
	}

	parts := strings.Split(raw, ".")

	return strings.Join(parts[:len(parts)-1], "."), nil
}

func (evt SubscriptionEvent) eventResourcePrefix() (string, error) {
	r, err := evt.splitEventType()

	return r, err
}

func resourceRecord(evt SubscriptionEvent) (objectName string, record map[string]any, err error) {
	resource, err := evt.splitEventType()
	if err != nil {
		return "", nil, err
	}

	payloadObject, connectorObject, err := resolveEventObject(resource)
	if err != nil {
		return "", nil, err
	}

	v, ok := evt[payloadObject].(map[string]any)
	if !ok || v == nil {
		return "", nil, fmt.Errorf("%w: missing payload %q", errMalformedWebhookEvent, payloadObject)
	}

	return connectorObject, v, nil
}

func resolveEventObject(resource string) (payloadObject string, connectorObject string, err error) {
	// job appointment and estimate option are not supported.
	if resource == "job.appointment" || resource == "estimate.option" {
		return "", "", fmt.Errorf("%w: %q", errUnsupportedWebhookResource, resource)
	}

	// event name can be object.action or object.subtype.action.
	payloadObject = strings.Split(resource, ".")[0]
	connectorObject = naming.NewSingularString(payloadObject).Plural().String()

	if supportedEventObjects.Has(connectorObject) {
		return payloadObject, connectorObject, nil
	}

	return "", "", fmt.Errorf("%w: %q", errUnsupportedWebhookResource, resource)
}
