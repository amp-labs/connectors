package calendly

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

var (
	_ common.SubscriptionEvent = CalendlyWebhookEvent{}
	
	ErrMissingSignature = errors.New("missing signature in webhook request")
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrMissingTimestamp = errors.New("missing timestamp in webhook request")
	ErrInvalidTimestamp = errors.New("invalid timestamp in webhook request")
)

// CalendlyWebhookEvent represents a webhook event from Calendly
type CalendlyWebhookEvent map[string]any

// CalendlyVerificationParams contains the parameters needed to verify Calendly webhook signatures
type CalendlyVerificationParams struct {
	SigningKey string `json:"signing_key"`
}

// Calendly webhook headers
const (
	calendlySignatureHeader = "Calendly-Webhook-Signature"
	calendlyTimestampHeader = "Calendly-Webhook-Timestamp"
)

// VerifyWebhookMessage verifies the signature of a webhook message from Calendly
func (c *Connector) VerifyWebhookMessage( //nolint:funlen
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if params.Param == nil {
		// If no verification params provided, accept the webhook (for testing)
		return true, nil
	}

	calendlyParams, err := common.AssertType[*CalendlyVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("invalid verification params: %w", err)
	}

	if calendlyParams.SigningKey == "" {
		// If no signing key provided, accept the webhook (for testing)
		return true, nil
	}

	// Get signature from headers
	signature := request.Headers.Get(calendlySignatureHeader)
	if signature == "" {
		return false, ErrMissingSignature
	}

	// Get timestamp from headers
	timestampStr := request.Headers.Get(calendlyTimestampHeader)
	if timestampStr == "" {
		return false, ErrMissingTimestamp
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, ErrInvalidTimestamp
	}

	stringToSign := timestampStr + "." + string(request.Body)

	mac := hmac.New(sha256.New, []byte(calendlyParams.SigningKey))
	mac.Write([]byte(stringToSign))
	expectedMac := mac.Sum(nil)

	expectedSignature := base64.StdEncoding.EncodeToString(expectedMac)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return false, ErrInvalidSignature
	}

	// Check timestamp to prevent replay attacks (within 5 minutes)
	now := time.Now().Unix()
	if now-timestamp > 300 {
		return false, fmt.Errorf("webhook timestamp too old")
	}

	return true, nil
}

// EventType returns the type of subscription event
func (e CalendlyWebhookEvent) EventType() (common.SubscriptionEventType, error) {
	eventName, err := e.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, err
	}

	switch eventName {
	case "invitee.created":
		return common.SubscriptionEventTypeCreate, nil
	case "invitee.canceled":
		return common.SubscriptionEventTypeDelete, nil
	case "invitee_no_show.created", "invitee_no_show.deleted":
		return common.SubscriptionEventTypeOther, nil
	case "routing_form_submission.created":
		return common.SubscriptionEventTypeOther, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

// RawEventName returns the raw event name from the webhook payload
func (e CalendlyWebhookEvent) RawEventName() (string, error) {
	event, ok := e["event"]
	if !ok {
		return "", fmt.Errorf("missing event field in webhook payload")
	}

	eventStr, ok := event.(string)
	if !ok {
		return "", fmt.Errorf("event field is not a string")
	}

	return eventStr, nil
}

// ObjectName returns the object name for the event (always "scheduled_events" for Calendly)
func (e CalendlyWebhookEvent) ObjectName() (string, error) {
	return "scheduled_events", nil
}

// Workspace returns the workspace/organization identifier
func (e CalendlyWebhookEvent) Workspace() (string, error) {
	payload, ok := e["payload"]
	if !ok {
		return "", nil
	}

	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return "", nil
	}

	if eventUri, ok := payloadMap["event"].(map[string]any); ok {
		if uri, ok := eventUri["uri"].(string); ok {
			return e.extractOrganizationFromURI(uri), nil
		}
	}

	return "", nil
}

// RecordId returns the record ID from the event payload
func (e CalendlyWebhookEvent) RecordId() (string, error) {
	payload, ok := e["payload"]
	if !ok {
		return "", fmt.Errorf("missing payload in webhook event")
	}

	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return "", fmt.Errorf("payload is not a map")
	}

	if invitee, ok := payloadMap["invitee"].(map[string]any); ok {
		if uri, ok := invitee["uri"].(string); ok {
			return e.extractIDFromURI(uri), nil
		}
	}

	if event, ok := payloadMap["event"].(map[string]any); ok {
		if uri, ok := event["uri"].(string); ok {
			return e.extractIDFromURI(uri), nil
		}
	}

	return "", fmt.Errorf("could not extract record ID from payload")
}

// EventTimeStampNano returns the event timestamp in nanoseconds
func (e CalendlyWebhookEvent) EventTimeStampNano() (int64, error) {
	timeStr, ok := e["time"]
	if !ok {
		return 0, fmt.Errorf("missing time field in webhook event")
	}

	timeString, ok := timeStr.(string)
	if !ok {
		return 0, fmt.Errorf("time field is not a string")
	}

	// Parse the time string (RFC3339 format)
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse time: %w", err)
	}

	return t.UnixNano(), nil
}

// extractIDFromURI extracts the ID from a Calendly URI
func (e CalendlyWebhookEvent) extractIDFromURI(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return uri
}

// extractOrganizationFromURI attempts to extract organization info from a URI
func (e CalendlyWebhookEvent) extractOrganizationFromURI(uri string) string {
	return ""
} 