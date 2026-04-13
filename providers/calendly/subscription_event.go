package calendly

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

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Compile-time interface checks.
var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}
)

const headerCalendlyWebhookSignature = "Calendly-Webhook-Signature"

// SubscriptionEvent is a Calendly webhook JSON body (map shape — see Calendly webhook payload docs).
type SubscriptionEvent map[string]any

// CalendlyVerificationParams carries the signing key used when the webhook subscription was created.
type CalendlyVerificationParams struct {
	SigningKey string
}

func (evt SubscriptionEvent) PreLoadData(*common.SubscriptionEventPreLoadData) error {
	return nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := common.StringMap(evt)

	s, err := m.GetString("event")
	if err != nil {
		return "", fmt.Errorf("calendly webhook: %w", err)
	}

	return s, nil
}

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	raw, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, err
	}

	parts := splitEventName(raw)
	if len(parts) != 2 { //nolint:mnd
		return common.SubscriptionEventTypeOther, fmt.Errorf("unexpected event name: %q", raw)
	}

	switch parts[1] {
	case "created":
		return common.SubscriptionEventTypeCreate, nil
	case "updated", "propertyChange":
		return common.SubscriptionEventTypeUpdate, nil
	case "deleted", "canceled":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	raw, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	parts := splitEventName(raw)
	if len(parts) != 2 { //nolint:mnd
		return "", fmt.Errorf("unexpected event name: %q", raw)
	}

	switch parts[0] {
	case "event_type":
		return "event_types", nil
	case "invitee", "invitee_no_show":
		return "scheduled_events", nil
	case "routing_form_submission":
		return "routing_forms", nil
	default:
		return "", fmt.Errorf("unsupported webhook event family: %q", parts[0])
	}
}

func (evt SubscriptionEvent) Workspace() (string, error) {
	root := common.StringMap(evt)

	if org, err := root.GetString("organization"); err == nil && org != "" {
		return org, nil
	}

	pm, err := evt.payloadMap()
	if err != nil {
		return "", err
	}

	if org, err := pm.GetString("organization"); err == nil && org != "" {
		return org, nil
	}

	if org, err := nestedString(pm, "scheduled_event", "organization"); err == nil && org != "" {
		return org, nil
	}

	return "", errors.New("calendly webhook: organization not found")
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	pm, err := evt.payloadMap()
	if err != nil {
		return "", err
	}

	// Event type webhooks: resource is the event type URI.
	if uri, err := pm.GetString("event_type"); err == nil && uri != "" {
		return uri, nil
	}

	if uri, err := pm.GetString("uri"); err == nil && uri != "" {
		return uri, nil
	}

	if uri, err := nestedString(pm, "event_type", "uri"); err == nil && uri != "" {
		return uri, nil
	}

	if uri, err := nestedString(pm, "scheduled_event", "uri"); err == nil && uri != "" {
		return uri, nil
	}

	return "", errors.New("calendly webhook: record uri not found in payload")
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	pm, err := evt.payloadMap()
	if err != nil {
		return 0, err
	}

	// Try common timestamp fields.
	for _, key := range []string{"created_at", "updated_at", "canceled_at"} {
		if s, err := pm.GetString(key); err == nil && s != "" {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t.UnixNano(), nil
			}
		}
	}

	return 0, errors.New("calendly webhook: timestamp not found in payload")
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	return nil, nil
}

func (evt SubscriptionEvent) payloadMap() (common.StringMap, error) {
	m := common.StringMap(evt)

	p, err := m.Get("payload")
	if err != nil {
		// Some samples use a flat body; treat the root object as the payload.
		return m, nil
	}

	pm, ok := p.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("calendly webhook: payload is %T, expected object", p)
	}

	return common.ToStringMap(pm), nil
}

func nestedString(m common.StringMap, path ...string) (string, error) {
	var cur any = map[string]any(m)

	for _, seg := range path {
		mm, ok := cur.(map[string]any)
		if !ok {
			return "", errNestedWalk
		}

		v, ok := mm[seg]
		if !ok {
			return "", errNestedWalk
		}

		cur = v
	}

	s, ok := cur.(string)
	if !ok {
		return "", errNestedWalk
	}

	return s, nil
}

var errNestedWalk = errors.New("nested field not found")

// VerifyWebhookMessage verifies Calendly's webhook signature (Calendly-Webhook-Signature header).
// Format: t=<unix_seconds>,v1=<hex_hmac> per https://developer.calendly.com/api-docs/ZG9jOjM2MzE2MDM4-webhook-signatures
func (*Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	p, err := common.AssertType[*CalendlyVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("calendly: invalid verification params: %w", err)
	}

	if p.SigningKey == "" {
		return false, errors.New("calendly: signing key is empty")
	}

	sigHeader := request.Headers.Get(headerCalendlyWebhookSignature)
	if sigHeader == "" {
		return false, errors.New("calendly: missing Calendly-Webhook-Signature header")
	}

	ts, v1, err := parseCalendlySignatureHeader(sigHeader)
	if err != nil {
		return false, err
	}

	expected := calendlyHMACHex(p.SigningKey, ts, request.Body)

	ok, err := secureCompareHex(expected, v1)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func parseCalendlySignatureHeader(header string) (timestamp, v1 string, err error) {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2) //nolint:mnd
		if len(kv) != 2 {                    //nolint:mnd
			continue
		}

		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}

	if timestamp == "" || v1 == "" {
		return "", "", errors.New("calendly: signature header missing t= or v1=")
	}

	return timestamp, v1, nil
}

func calendlyHMACHex(signingKey, timestamp string, body []byte) string {
	// t + "." + raw_body (UTF-8) — common pattern documented for Calendly webhooks.
	msg := timestamp + "." + string(body)

	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(msg))

	return hex.EncodeToString(mac.Sum(nil))
}

func secureCompareHex(expectedHex, received string) (bool, error) {
	received = strings.TrimSpace(strings.ToLower(received))
	expectedHex = strings.TrimSpace(strings.ToLower(expectedHex))

	eb, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false, fmt.Errorf("calendly: decode expected hex: %w", err)
	}

	rb, err := hex.DecodeString(received)
	if err != nil {
		return false, fmt.Errorf("calendly: decode signature hex: %w", err)
	}

	if len(eb) == 0 || len(rb) == 0 {
		return false, errors.New("calendly: empty signature bytes")
	}

	return hmac.Equal(eb, rb), nil
}

var _ connectors.WebhookVerifierConnector = (*Connector)(nil)
