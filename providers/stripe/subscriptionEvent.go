package stripe

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

const (
	minEventTypeParts = 2
)

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}

	errTypeMismatch = errors.New("type mismatch")
)

// SubscriptionEvent represents a webhook event from Stripe.
type SubscriptionEvent map[string]any

// CollapsedSubscriptionEvent represents the raw webhook payload from Stripe.
// Stripe sends one event per webhook, so this implementation simply wraps the single event.
type CollapsedSubscriptionEvent map[string]any

var _ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}

// RawMap returns a copy of the raw event data.
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	m := evt.asMap()

	data, err := m.Get("data")
	if err != nil {
		return nil, err
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected map[string]any, got %T", errTypeMismatch, data)
	}

	previousAttrs, ok := dataMap["previous_attributes"].(map[string]any)
	if !ok {
		return []string{}, nil
	}

	updatedFields := make([]string, 0, len(previousAttrs))
	for field := range previousAttrs {
		updatedFields = append(updatedFields, field)
	}

	return updatedFields, nil
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	m := evt.asMap()

	created, err := m.AsInt("created")
	if err != nil {
		return 0, err
	}

	// Stripe timestamps are in seconds, convert to nanoseconds
	return time.Unix(created, 0).UnixNano(), nil
}

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	eventType, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	parts := strings.Split(eventType, ".")
	if len(parts) < minEventTypeParts {
		return common.SubscriptionEventTypeOther, nil
	}

	action := strings.ToLower(parts[len(parts)-1])

	switch action {
	case "created":
		return common.SubscriptionEventTypeCreate, nil
	case "updated":
		return common.SubscriptionEventTypeUpdate, nil
	case "deleted":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	m := evt.asMap()

	data, err := m.Get("data")
	if err != nil {
		return evt.extractObjectNameFromEventType()
	}

	dataMap, isDataMap := data.(map[string]any)
	if !isDataMap {
		return evt.extractObjectNameFromEventType()
	}

	obj, isObjMap := dataMap["object"].(map[string]any)
	if !isObjMap {
		return evt.extractObjectNameFromEventType()
	}

	objectType, isString := obj["object"].(string)
	if !isString || objectType == "" {
		return evt.extractObjectNameFromEventType()
	}

	return objectType, nil
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	eventType, err := m.GetString("type")
	if err != nil {
		return "", err
	}

	return eventType, nil
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

	dataMap, isDataMap := data.(map[string]any)
	if !isDataMap {
		return "", fmt.Errorf("%w: expected map[string]any, got %T", errTypeMismatch, data)
	}

	obj, isObjMap := dataMap["object"].(map[string]any)
	if !isObjMap {
		return "", fmt.Errorf("%w: expected map[string]any for data.object, got %T", errTypeMismatch, dataMap["object"])
	}

	id, isString := obj["id"].(string)
	if !isString {
		return "", fmt.Errorf("%w: expected string for data.object.id, got %T", errTypeMismatch, obj["id"])
	}

	return id, nil
}

// Workspace returns an empty string as Stripe doesn't have a workspace concept.
func (evt SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

// extractObjectNameFromEventType extracts the object name from the event type.
// Example: "setup_intent.created" -> "setup_intent".
func (evt SubscriptionEvent) extractObjectNameFromEventType() (string, error) {
	eventType, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	parts := strings.Split(eventType, ".")
	if len(parts) < minEventTypeParts {
		return "", fmt.Errorf("%w: %s", errInvalidEventTypeFormat, eventType)
	}

	objectName := strings.Join(parts[:len(parts)-1], ".")

	return objectName, nil
}

// asMap returns the event as a StringMap.
func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
}
