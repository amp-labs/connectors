package webhook

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Event is a singular notification message within EventCollection.
type Event map[string]any

var (
	_ common.SubscriptionEvent       = Event{}
	_ common.SubscriptionUpdateEvent = Event{}
)

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

	identifierHolder, ok := objectNameToIdentifierHolder[objectName] // nolint:varnamelen
	if !ok {
		return "", fmt.Errorf("unknown object name %v", objectName) // nolint:err113
	}

	eventEntity, ok := e["event"]
	if !ok {
		return "", errors.New("missing 'event' key") // nolint:err113
	}

	eventObject, ok := eventEntity.(map[string]any)
	if !ok {
		return "", errors.New("'event' is not of type map[string]any") // nolint:err113
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
var objectNameToIdentifierHolder = datautils.Map[string, string]{ // nolint:gochecknoglobals
	"bots":          "bot",
	"calls":         "call",
	"conversations": "channel",
	"files":         "file",
	"users":         "user",
}
