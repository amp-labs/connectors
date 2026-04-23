package google

import (
	"maps"
	"time"

	"github.com/amp-labs/connectors/common"
)

// Gmail history change categories, as they appear in the history.list response.
// Stored verbatim on SubscriptionEvent.RawName so EventType can map them to the
// common subscription event taxonomy.
const (
	historyCategoryMessagesAdded   = "messagesAdded"
	historyCategoryMessagesDeleted = "messagesDeleted"
	historyCategoryLabelsAdded     = "labelsAdded"
	historyCategoryLabelsRemoved   = "labelsRemoved"
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
	MessageID    string `json:"messageId"    mapstructure:"messageId"`
	HistoryID    string `json:"historyId"    mapstructure:"historyId"`
	EmailAddress string `json:"emailAddress" mapstructure:"emailAddress"`
	// RawName is the Gmail history change category (e.g. "messagesAdded").
	RawName    string `json:"rawEventName"         mapstructure:"rawEventName"`
	OccurredAt int64  `json:"occurredAt,omitempty" mapstructure:"occurredAt"`
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

// EventType maps the Gmail history change category stored in RawName to the
// common subscription event taxonomy.
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	switch e.RawName {
	case historyCategoryMessagesAdded:
		return common.SubscriptionEventTypeCreate, nil
	case historyCategoryMessagesDeleted:
		return common.SubscriptionEventTypeDelete, nil
	case historyCategoryLabelsAdded, historyCategoryLabelsRemoved:
		return common.SubscriptionEventTypeUpdate, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

// RawEventName returns the Gmail history change category verbatim
// (e.g. "messagesAdded", "labelsRemoved").
func (e SubscriptionEvent) RawEventName() (string, error) {
	return e.RawName, nil
}

// ObjectName returns the Gmail object name. Only "messages" is supported for
// subscriptions today.
//
// TODO(ENG-3851): derive the object name by parsing RawName (e.g.
// "messagesAdded" → "messages", "labelsAdded" → "labels") rather than
// hardcoding. https://linear.app/ampersand/issue/ENG-3851
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
		"rawEventName": e.RawName,
		"occurredAt":   e.OccurredAt,
	}), nil
}

// PreLoadData is a no-op for Gmail — the event carries everything it needs
// at construction time. Kept to satisfy the common.SubscriptionEvent interface.
func (e SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

// SubscriptionEventsFromHistory fans a Gmail history.list response out into one
// SubscriptionEvent per affected message. Each event stores the Gmail-native
// change category (e.g. "messagesAdded") in RawName; callers use EventType()
// to map that into the common.SubscriptionEventType taxonomy.
//
// When the same message id appears in multiple categories within a single
// history fetch, the last-seen category wins — matching the "last write"
// semantics the server-side dispatch pipeline expects (one entry per record id).
//
// The emailID argument populates every event's Workspace/EmailAddress so the
// subscribe pipeline can route each event back to the right connection.
func SubscriptionEventsFromHistory(history []HistoryRecord, emailID string) []SubscriptionEvent {
	seen := make(map[string]int)

	var events []SubscriptionEvent

	add := func(msgID, historyID, rawName string) {
		if idx, ok := seen[msgID]; ok {
			events[idx].RawName = rawName
			events[idx].HistoryID = historyID

			return
		}

		seen[msgID] = len(events)
		events = append(events, SubscriptionEvent{
			MessageID:    msgID,
			HistoryID:    historyID,
			EmailAddress: emailID,
			RawName:      rawName,
		})
	}

	for _, rec := range history {
		for _, added := range rec.MessagesAdded {
			add(added.Message.ID, rec.ID, historyCategoryMessagesAdded)
		}

		for _, deleted := range rec.MessagesDeleted {
			add(deleted.Message.ID, rec.ID, historyCategoryMessagesDeleted)
		}

		for _, labelAdded := range rec.LabelsAdded {
			add(labelAdded.Message.ID, rec.ID, historyCategoryLabelsAdded)
		}

		for _, labelRemoved := range rec.LabelsRemoved {
			add(labelRemoved.Message.ID, rec.ID, historyCategoryLabelsRemoved)
		}
	}

	return events
}
