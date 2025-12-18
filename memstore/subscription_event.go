package memstore

import (
	"time"

	"github.com/amp-labs/connectors/common"
)

// SubscriptionEvent implements the common.SubscriptionEvent interface for memstore test events.
type SubscriptionEvent struct {
	// EventTypeValue is the type of event (create, update, delete)
	EventTypeValue common.SubscriptionEventType `mapstructure:"eventType"`
	// ObjectNameValue is the object name (e.g., "accounts", "contacts")
	ObjectNameValue string `mapstructure:"objectName"`
	// RecordIDValue is the ID of the affected record
	RecordIDValue string `mapstructure:"recordId"`
	// EventTimeValue is when the event occurred
	EventTimeValue int64 `mapstructure:"eventTime"`
	// RawData contains the raw event data
	RawData map[string]any `mapstructure:",remain"`
}

// EventType returns the type of event (create, update, delete).
func (e *SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	if e.EventTypeValue == "" {
		return "", ErrEventTypeEmpty
	}

	return e.EventTypeValue, nil
}

// RawEventName returns the raw event name from the provider.
// For memstore, this is the same as EventType.
func (e *SubscriptionEvent) RawEventName() (string, error) {
	return string(e.EventTypeValue), nil
}

// ObjectName returns the object name (e.g., "accounts", "contacts").
func (e *SubscriptionEvent) ObjectName() (string, error) {
	if e.ObjectNameValue == "" {
		return "", ErrObjectNameEmpty
	}

	return e.ObjectNameValue, nil
}

// Workspace returns the workspace identifier.
// For memstore, we don't have workspaces, so return empty string.
func (e *SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

// RecordId returns the ID of the affected record.
func (e *SubscriptionEvent) RecordId() (string, error) {
	if e.RecordIDValue == "" {
		return "", ErrRecordIDEmpty
	}

	return e.RecordIDValue, nil
}

// EventTimeStampNano returns the event timestamp in nanoseconds.
func (e *SubscriptionEvent) EventTimeStampNano() (int64, error) {
	if e.EventTimeValue == 0 {
		// If no time provided, use current time
		return time.Now().UnixNano(), nil
	}

	return e.EventTimeValue, nil
}

// RawMap returns the raw event data as a map.
func (e *SubscriptionEvent) RawMap() (map[string]any, error) {
	if e.RawData == nil {
		return map[string]any{
			"eventType":  e.EventTypeValue,
			"objectName": e.ObjectNameValue,
			"recordId":   e.RecordIDValue,
			"eventTime":  e.EventTimeValue,
		}, nil
	}

	return e.RawData, nil
}
