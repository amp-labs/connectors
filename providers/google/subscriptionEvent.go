package google

import (
	"maps"
	"time"

	"github.com/amp-labs/connectors/common"
)

// SubscriptionEvent is a single Gmail message-level event derived from a Gmail
// watch push notification. Gmail pushes only carry {emailAddress, historyId};
// the caller (typically a worker that received the push) is expected to call
// HistoryList, expand the result into one SubscriptionEvent per affected
// message, and republish the expanded events through the standard subscribe
// pipeline so they can be hydrated and delivered like any other provider's
// events.
//
// This type is a struct (rather than a map like HubSpot's SubscriptionEvent)
// because Ampersand generates these events itself from HistoryList output,
// so the shape is fully known and stable.
// mapstructure tags mirror the json tags because the server's messenger
// roundtrips this struct through mapstructure (both decoding the incoming
// webhook body and re-emitting the event as a map for RawEvent). Without
// tags on every field, that re-emit leaks Go field names into the payload.
type SubscriptionEvent struct {
	MessageID    string                       `json:"messageId"              mapstructure:"messageId"`
	HistoryID    string                       `json:"historyId"              mapstructure:"historyId"`
	EmailAddress string                       `json:"emailAddress"           mapstructure:"emailAddress"`
	Type         common.SubscriptionEventType `json:"eventType"              mapstructure:"eventType"`
	OccurredAt   int64                        `json:"occurredAt,omitempty"   mapstructure:"occurredAt"`
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

// EventType returns the subscription event category (create/update/delete).
// Falls back to Other if the type field is empty, matching the contract of
// other providers' SubscriptionEvent implementations.
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	if e.Type == "" {
		return common.SubscriptionEventTypeOther, nil
	}

	return e.Type, nil
}

// RawEventName returns the event type as a string. Gmail has no separate
// provider-native event name (unlike HubSpot's "contact.creation"), so this
// mirrors EventType.
func (e SubscriptionEvent) RawEventName() (string, error) {
	return string(e.Type), nil
}

// ObjectName returns the Gmail object name. Only "messages" is supported for
// subscriptions today.
func (e SubscriptionEvent) ObjectName() (string, error) {
	return "messages", nil
}

// Workspace returns the Gmail mailbox address. The subscribe pipeline uses
// this to route events to the right connection — Gmail connections store the
// mailbox as provider_consumer_ref.
func (e SubscriptionEvent) Workspace() (string, error) {
	return e.EmailAddress, nil
}

// RecordId returns the Gmail message ID, used by the subscribe pipeline to
// hydrate the full message via GetRecordsByIds.
func (e SubscriptionEvent) RecordId() (string, error) {
	return e.MessageID, nil
}

// EventTimeStampNano returns the event's occurrence time in nanoseconds.
// Falls back to the time of this call when OccurredAt is unset, since Gmail
// history records don't carry a per-message timestamp.
func (e SubscriptionEvent) EventTimeStampNano() (int64, error) {
	if e.OccurredAt != 0 {
		return e.OccurredAt, nil
	}

	return time.Now().UnixNano(), nil
}

// RawMap returns a map representation of the event. The subscribe pipeline
// decodes this into entry.RawEvent on outgoing webhooks.
func (e SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(map[string]any{
		"messageId":    e.MessageID,
		"historyId":    e.HistoryID,
		"emailAddress": e.EmailAddress,
		"eventType":    string(e.Type),
		"occurredAt":   e.OccurredAt,
	}), nil
}

// PreLoadData is a no-op for Gmail — the event carries everything it needs
// at construction time. Kept to satisfy the common.SubscriptionEvent interface.
func (e SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}
