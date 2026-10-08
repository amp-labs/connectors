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

// createdUpdatedEpsilon is how close an event's updated time must be to its created time to
// classify as a create. On a genuine create the two are near-identical, but not equal: created
// is truncated to whole seconds while updated keeps milliseconds (live-observed delta: 318ms).
// 2s absorbs that skew while excluding human edits-after-create. Trade-off: an async
// post-create touch by Google that bumps updated past 2s (e.g. Meet link resolution) would
// deliver that create as an update.
const createdUpdatedEpsilon = 2 * time.Second

// SubscriptionEvent is one Calendar event to be classified, built from a GetRecordsByIds row.
//
// Calendar pushes have an empty body, so the subscribe pipeline fetches the changed events
// first and classifies them afterwards. Keeping that split, GetRecordsByIds is a plain read
// and EventType does the classification here. A Calendar event has no event-type field (Gmail
// gets one from its history categories), so the type is inferred from the status, the created
// time relative to the fetch window (UpdatedMin, the checkpoint GetRecordsByIds used), and how
// closely updated trails created.
type SubscriptionEvent struct {
	// RecordID is the Calendar event ID.
	RecordID string
	// Status is the event status; "cancelled" means the event was deleted.
	Status string
	// Created is the event creation time (RFC3339). EventType compares it against UpdatedMin
	// and Updated to tell a new event from an edit to an existing one.
	Created string
	// Updated is the last-modification time (RFC3339), used as the event timestamp and, with
	// Created, to classify creates (a genuine create has updated within a few seconds of
	// created).
	Updated string
	// UpdatedMin is the fetch window checkpoint (recordIds[0] passed to GetRecordsByIds).
	UpdatedMin string
	// Raw is the full event object from the provider.
	Raw map[string]any
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

// EventType infers the event type from the row. A cancelled event is a delete; otherwise an
// event is a create only when it was created within the fetch window AND its updated time sits
// within createdUpdatedEpsilon of its created time. Everything else is an update.
//
// The window check alone is not enough: the server's updatedMin checkpoint lags real time by a
// couple of minutes (lookback buffer plus time between pushes), so an event edited shortly
// after creation still has created >= updatedMin and would re-classify as a create. On a
// genuine create updated barely trails created, while any edit pushes updated well past it, so
// the epsilon is what separates the two. The window check stays as a guard against backfills
// that rewrite an old event's updated time to near its created time.
//
// updatedMin defines the window and must parse, or we error. A non-cancelled event with no
// usable created time falls back to update; one created in the window with no usable updated
// time falls back to create (the pre-epsilon behavior).
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

	// Created before the window start means it existed before this fetch: an edit.
	if created.Before(window) {
		return common.SubscriptionEventTypeUpdate, nil
	}

	updated, err := time.Parse(time.RFC3339, e.Updated)
	if err != nil {
		// Created in the window but nothing to measure the edit gap with; call it a create.
		return common.SubscriptionEventTypeCreate, nil // nolint: nilerr
	}

	if updated.Sub(created) <= createdUpdatedEpsilon {
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
// if Updated is missing or unparsable.
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
