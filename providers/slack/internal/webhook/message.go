package webhook

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

// CollapsedSubscriptionEvent represents the raw webhook payload.
// This simply wraps the single event.
type CollapsedSubscriptionEvent map[string]any

// Event is a singular notification message.
type Event map[string]any

var (
	_ common.SubscriptionEvent          = Event{}
	_ common.SubscriptionUpdateEvent    = Event{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
)

func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	event := Event(e)

	// Message events carry the full message inline in the payload and Slack has no
	// endpoint to fetch a single message by id, so they are wrapped as MessageEvent
	// to expose the inline record. Errors are ignored here: an unmapped event name
	// is reported by EventType/ObjectName during downstream processing.
	if objectName, err := event.ObjectName(); err == nil && objectName == objectNameMessages {
		return []common.SubscriptionEvent{MessageEvent{Event: event}}, nil
	}

	return []common.SubscriptionEvent{event}, nil
}

func (e Event) EventType() (common.SubscriptionEventType, error) {
	eventName, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	description, ok := eventNameToEventDescription[eventName]
	if !ok {
		return "", fmt.Errorf("eventName '%v' is unknown", eventName) // nolint:err113
	}

	return description.Type, nil
}

func (e Event) RawEventName() (string, error) {
	eventEntity, ok := e["event"] // nolint:varnamelen
	if !ok {
		return "", errors.New("missing 'event' key") // nolint:err113
	}

	eventObject, ok := eventEntity.(map[string]any)
	if !ok {
		return "", errors.New("'event' is not of type map[string]any") // nolint:err113
	}

	eventTypeEntity, ok := eventObject["type"]
	if !ok {
		return "", errors.New("'event' is missing key 'type'") // nolint:err113
	}

	eventType, ok := eventTypeEntity.(string)
	if !ok {
		return "", errors.New("'type' is not of type string") // nolint:err113
	}

	// Message events all share the type "message" and are differentiated by an
	// optional "subtype" field, so the subtype is the effective event name. Subtypes
	// not present in eventNameToEventDescription (bot_message, channel_join, ...) are
	// reported as unknown by EventType/ObjectName, like any unmapped event name.
	// A plain new message — thread replies included — carries no subtype. (Slack's
	// "message_replied" subtype for the parent is famously never populated; replies
	// are detected by their own "message" event carrying a "thread_ts".)
	if eventType == rawEventMessage {
		if subtype, ok := eventObject["subtype"].(string); ok && subtype != "" {
			return subtype, nil
		}
	}

	return eventType, nil
}

func (e Event) ObjectName() (string, error) {
	eventName, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	description, ok := eventNameToEventDescription[eventName]
	if !ok {
		return "", fmt.Errorf("eventName '%v' is unknown", eventName) // nolint:err113
	}

	return description.ObjectName, nil
}

// Workspace
// "team_id" property:
// > The unique identifier for the workspace/team where this event occurred. Example: T461EG9ZZ.
func (e Event) Workspace() (string, error) {
	teamIDEntity, ok := e["team_id"]
	if !ok {
		return "", errors.New("missing 'team_id' key") // nolint:err113
	}

	teamID, ok := teamIDEntity.(string)
	if !ok {
		return "", errors.New("'team_id' is not of type string") // nolint:err113
	}

	return teamID, nil
}

func (e Event) RecordId() (string, error) {
	objectName, err := e.ObjectName()
	if err != nil {
		return "", err
	}

	if objectName == objectNameMessages {
		return e.messageRecordId()
	}

	identifierHolder, ok := objectNameToIdentifierHolder[objectName] // nolint:varnamelen
	if !ok {
		return "", fmt.Errorf("unknown object name %v", objectName) // nolint:err113
	}

	eventObject, err := e.eventObject()
	if err != nil {
		return "", err
	}

	holder, ok := eventObject[identifierHolder]
	if !ok {
		return "", fmt.Errorf("'event' does not have '%v' key", identifierHolder) // nolint:err113
	}

	identifier, ok := holder.(string)
	if ok {
		return identifier, nil
	}

	// The identifier is nested.
	object, ok := holder.(map[string]any)
	if !ok {
		return "", fmt.Errorf("'event.%v' is not of type map[string]any", identifierHolder) // nolint:err113
	}

	identifierEntity, ok := object["id"]
	if !ok {
		return "", fmt.Errorf("'event.%v' does not have 'id' key", identifierHolder) // nolint:err113
	}

	identifier, ok = identifierEntity.(string)
	if !ok {
		return "", fmt.Errorf("'event.%v.id' is not of type string", identifierHolder) // nolint:err113
	}

	return identifier, nil
}

// EventTimeStampNano
// "event_time" property:
// > The epoch timestamp in seconds indicating when this event was dispatched.
func (e Event) EventTimeStampNano() (int64, error) {
	eventTimeEntity, ok := e["event_time"]
	if !ok {
		return 0, errors.New("missing 'event_time' key") // nolint:err113
	}

	eventTime, ok := eventTimeEntity.(float64)
	if !ok {
		return 0, errors.New("'event_time' is not of type float64") // nolint:err113
	}

	// Convert seconds to nanoseconds.
	seconds := int64(eventTime)

	return time.Unix(seconds, 0).UnixNano(), nil
}

func (e Event) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e Event) PreLoadData(data *common.SubscriptionEventPreLoadData) error {
	return nil
}

func (e Event) UpdatedFields() ([]string, error) {
	return nil, nil
}

// objectNameToIdentifierHolder maps Slack object names to the identifier holder field name
// used in the event payload. The identifier is located under one of two paths:
//   - "event.<identifier-holder>.id" — if the holder is an object (e.g., "event.bot.id")
//   - "event.<identifier-holder>" — if the ID is directly the value (e.g., "event.channel")
//
// Messages are absent on purpose: they have no standalone id (see messageRecordId).
var objectNameToIdentifierHolder = datautils.Map[string, string]{ // nolint:gochecknoglobals
	"bots":          "bot",
	"calls":         "call",
	"conversations": "channel",
	"files":         "file",
	"users":         "user",
}

func (e Event) eventObject() (map[string]any, error) {
	eventEntity, ok := e["event"]
	if !ok {
		return nil, errors.New("missing 'event' key") // nolint:err113
	}

	eventObject, ok := eventEntity.(map[string]any)
	if !ok {
		return nil, errors.New("'event' is not of type map[string]any") // nolint:err113
	}

	return eventObject, nil
}

// messageRecordId builds the identifier of the message a message.* event refers to.
// Slack messages have no standalone id — a message is uniquely identified by its
// channel and ts pair — so the record id is composed as "<channel>:<ts>".
func (e Event) messageRecordId() (string, error) {
	eventObject, err := e.eventObject()
	if err != nil {
		return "", err
	}

	channel, ok := eventObject["channel"].(string)
	if !ok {
		return "", errors.New("'event.channel' is missing or not of type string") // nolint:err113
	}

	timestamp, err := e.messageTimestamp(eventObject)
	if err != nil {
		return "", err
	}

	return channel + ":" + timestamp, nil
}

// messageTimestamp returns the ts of the message a message.* event refers to. It lives
// in a different place per subtype:
//   - message, thread_broadcast: "event.ts" — the event is the new message itself
//   - message_changed: "event.message.ts" — "event.ts" is the time of the edit
//   - message_deleted: "event.deleted_ts" — "event.ts" is the time of the deletion
func (e Event) messageTimestamp(eventObject map[string]any) (string, error) {
	eventName, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	holder, timestampKey := eventObject, "ts"

	switch eventName {
	case rawEventMessageChanged:
		message, err := changedMessage(eventObject)
		if err != nil {
			return "", err
		}

		holder = message
	case rawEventMessageDeleted:
		timestampKey = "deleted_ts"
	}

	timestamp, ok := holder[timestampKey].(string)
	if !ok {
		return "", fmt.Errorf("'%v' is missing or not of type string", timestampKey) // nolint:err113
	}

	return timestamp, nil
}

// changedMessage returns the updated message a message_changed event carries under
// "event.message".
func changedMessage(eventObject map[string]any) (map[string]any, error) {
	message, ok := eventObject["message"].(map[string]any)
	if !ok {
		return nil, errors.New("'event.message' is missing or not of type map[string]any") // nolint:err113
	}

	return message, nil
}

// MessageEvent is a message.* notification. Slack ships the full message inline in the
// webhook payload and has no endpoint to fetch a single message by id, so the event
// implements common.SubscriptionEventWithRecord: the record is attached straight from
// the payload instead of a GetRecordsByIds fetch.
type MessageEvent struct {
	Event
}

var _ common.SubscriptionEventWithRecord = MessageEvent{}

// Record returns the message carried inline in the webhook payload as a ReadResultRow,
// marshaled the same way a read returns it: Raw holds the full message and Fields holds
// the requested subset (lowercased). For "message" and "thread_broadcast" the event
// object itself is the message; for "message_changed" the updated message lives under
// "event.message" (the channel is only on the envelope, so it is copied onto the
// record). "message_deleted" carries no record body — only the id is meaningful
// downstream — so it returns an empty row.
func (e MessageEvent) Record(fields []string) (common.ReadResultRow, error) {
	eventObject, err := e.eventObject()
	if err != nil {
		return common.ReadResultRow{}, err
	}

	eventName, err := e.RawEventName()
	if err != nil {
		return common.ReadResultRow{}, err
	}

	var record map[string]any

	switch eventName {
	case rawEventMessageChanged:
		message, err := changedMessage(eventObject)
		if err != nil {
			return common.ReadResultRow{}, err
		}

		record = maps.Clone(message)
		record["channel"] = eventObject["channel"]
	case rawEventMessageDeleted:
		return common.ReadResultRow{}, nil
	default:
		record = maps.Clone(eventObject)
	}

	recordId, err := e.messageRecordId()
	if err != nil {
		return common.ReadResultRow{}, err
	}

	marshaler := readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("ts"))

	rows, err := marshaler([]map[string]any{record}, fields)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	// A webhook carries exactly one message, so the marshaler returns exactly one row.
	// The row id is the composite "<channel>:<ts>" identifier, matching RecordId().
	row := rows[0]
	row.Id = recordId

	return row, nil
}
