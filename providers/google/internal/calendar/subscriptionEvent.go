package calendar

import (
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

// Deleted events come back from events.list with status "cancelled" when showDeleted=true,
// which is how GetRecordsByIds queries.
const statusCancelled = "cancelled"

// SubscriptionEvent is one Calendar event to be classified, built from a GetRecordsByIds row.
//
// Calendar pushes have an empty body, so the subscribe pipeline fetches the changed events
// first and classifies them afterwards. Keeping that split, GetRecordsByIds is a plain read
// and EventType does the classification here. A Calendar event has no event-type field (Gmail
// gets one from its history categories), so the type is inferred from the status and the
// created time relative to the fetch window (UpdatedMin, the checkpoint GetRecordsByIds used).
type SubscriptionEvent struct {
	// RecordID is the Calendar event ID.
	RecordID string
	// Status is the event status; "cancelled" means the event was deleted.
	Status string
	// Created is the event creation time (RFC3339), compared against UpdatedMin to tell a
	// new event from an edit to an existing one.
	Created string
	// Updated is the last-modification time (RFC3339), used as the event timestamp.
	Updated string
	// UpdatedMin is the fetch window checkpoint (recordIds[0] passed to GetRecordsByIds).
	UpdatedMin string
	// Raw is the full event object from the provider.
	Raw map[string]any
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

// EventType infers the event type from the row. A cancelled event is a delete; otherwise an
// event created within the fetch window is a create, and an older one is an update.
//
// updatedMin defines the window and must parse, or we error. A non-cancelled event with no
// usable created time falls back to update.
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	if strings.EqualFold(e.Status, statusCancelled) {
		return common.SubscriptionEventTypeDelete, nil
	}

	window, err := time.Parse(time.RFC3339, e.UpdatedMin)
	if err != nil {
		return common.SubscriptionEventTypeOther,
			fmt.Errorf("%w: updatedMin %q: %w", errInvalidTimestamp, e.UpdatedMin, err)
	}

	created, err := time.Parse(time.RFC3339, e.Created)
	if err != nil {
		// No created time to compare against, so treat it as an edit.
		return common.SubscriptionEventTypeUpdate, nil // nolint: nilerr
	}

	// Created at or after the window start means it's new within this fetch.
	if !created.Before(window) {
		return common.SubscriptionEventTypeCreate, nil
	}

	return common.SubscriptionEventTypeUpdate, nil
}

// RawEventName returns the event status (e.g. "confirmed", "cancelled"). Calendar has no
// event-name field, so status is the nearest thing.
func (e SubscriptionEvent) RawEventName() (string, error) {
	return e.Status, nil
}

// ObjectName returns the Calendar object name. Only "events" is supported for subscriptions.
func (e SubscriptionEvent) ObjectName() (string, error) {
	return objectNameEvents, nil
}

// Workspace doesn't apply to Calendar — there's no per-event mailbox like Gmail's address.
func (e SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

// RecordId returns the Calendar event ID, which the pipeline uses to hydrate the event.
func (e SubscriptionEvent) RecordId() (string, error) {
	return e.RecordID, nil
}

// EventTimeStampNano returns the last-modification time in nanoseconds, or the current time
// if Updated is missing or unparseable.
func (e SubscriptionEvent) EventTimeStampNano() (int64, error) {
	updated, err := time.Parse(time.RFC3339, e.Updated)
	if err != nil {
		return time.Now().UnixNano(), nil // nolint: nilerr
	}

	return updated.UnixNano(), nil
}

// RawMap returns the full event object for the outgoing webhook's RawEvent.
func (e SubscriptionEvent) RawMap() (map[string]any, error) {
	if e.Raw == nil {
		return map[string]any{}, nil
	}

	return maps.Clone(e.Raw), nil
}

// PreLoadData is a no-op; a Calendar event already holds everything it needs. It's here only
// to satisfy common.SubscriptionEvent.
func (e SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

// SubscriptionEventsFromRecords turns GetRecordsByIds rows into events to classify, tagging
// each with the fetch window so EventType can tell a new event from an edit. updatedMin must
// be the same checkpoint passed to GetRecordsByIds (recordIds[0]).
func SubscriptionEventsFromRecords(rows []common.ReadResultRow, updatedMin string) []SubscriptionEvent {
	events := make([]SubscriptionEvent, 0, len(rows))

	for _, row := range rows {
		events = append(events, SubscriptionEvent{
			RecordID:   rowString(row, "id", row.Id),
			Status:     rowString(row, "status", ""),
			Created:    rowString(row, "created", ""),
			Updated:    rowString(row, "updated", ""),
			UpdatedMin: updatedMin,
			Raw:        row.Raw,
		})
	}

	return events
}

// rowString reads a string field from a row, checking Raw first and then Fields (whose keys
// are lowercased). Returns fallback if neither has it.
func rowString(row common.ReadResultRow, key, fallback string) string {
	if v, ok := stringFromMap(row.Raw, key); ok {
		return v
	}

	if v, ok := stringFromMap(row.Fields, strings.ToLower(key)); ok {
		return v
	}

	return fallback
}

func stringFromMap(m map[string]any, key string) (string, bool) {
	if m == nil {
		return "", false
	}

	v, ok := m[key].(string)

	return v, ok && v != ""
}
