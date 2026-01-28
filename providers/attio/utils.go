package attio

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// objectEvents holds the mapping of create, update, and delete events for a specific Attio core object.
func getObjectEvents(objectName common.ObjectName) (objectEvents, bool) {
	events, exists := attioObjectEvents[objectName]
	if !exists {
		return objectEvents{}, false
	}

	return events, true
}

// getAllSupportedEvents returns all provider events that this object supports.
func (e objectEvents) getAllSupportedEvents() []providerEvent {
	var events []providerEvent

	events = append(events, e.createEvents...)
	events = append(events, e.updateEvents...)
	events = append(events, e.deleteEvents...)

	return events
}

// getProviderEvents converts a common event type to corresponding Attio provider events.
func (e objectEvents) toProviderEvents(commonEvent common.SubscriptionEventType) []providerEvent {
	switch commonEvent { // nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return e.createEvents
	case common.SubscriptionEventTypeUpdate:
		return e.updateEvents
	case common.SubscriptionEventTypeDelete:
		return e.deleteEvents
	default:
		return nil
	}
}

// toRecordEvents converts a common event type to generic Attio record events.
// This is used for standard/custom objects (people, companies, deals, etc.) which use
// generic "record.*" event types instead of object-specific events.
func toRecordEvents(commonEvent common.SubscriptionEventType) (providerEvent, error) {
	//nolint:exhaustive
	switch commonEvent {
	case common.SubscriptionEventTypeCreate:
		return "record.created", nil

	case common.SubscriptionEventTypeUpdate:
		return "record.updated", nil

	case common.SubscriptionEventTypeDelete:
		return "record.updated", nil

	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedSubscriptionEvent, commonEvent)
	}
}

func buildSubscriptionPayloadForCoreObj(
	objEvents objectEvents,
	objectName common.ObjectName,
	event common.SubscriptionEventType,
) (sub []subscription, err error) {
	providerEvents := objEvents.toProviderEvents(event)
	if len(providerEvents) == 0 {
		return nil, fmt.Errorf("%w: object '%s' with event '%s'",
			errUnsupportedSubscriptionEvent, objectName, event)
	}

	for _, e := range providerEvents {
		sub = append(sub, subscription{
			EventType: e,
			Filter:    nil,
		})
	}

	return sub, nil
}

func buildSubscriptionPayloadForStandardObj(
	standardObjects map[common.ObjectName]string,
	objectName common.ObjectName,
	event common.SubscriptionEventType,
) (sub []subscription, err error) {
	// Handle building subscriptions for standard/custom objects
	objectId, exists := standardObjects[objectName]
	if !exists {
		return nil, fmt.Errorf("%s: %w", objectName, errObjectNotFound)
	}

	recordEvent, err := toRecordEvents(event)
	if err != nil {
		return nil, err
	}

	// filter is required to specify the object type for standard/custom objects
	// It tells Attio which object (people, companies, deals, etc.) to subscribe to
	// since the event types are generic (record.created, record.updated, record.deleted)
	// Ref: https://docs.attio.com/rest-api/guides/webhooks#filtering
	filter := map[string]any{
		"$and": []map[string]any{
			{
				"field":    "id.object_id",
				"operator": "equals",
				"value":    objectId,
			},
		},
	}
	sub = append(sub,
		subscription{
			EventType: recordEvent,
			Filter:    filter,
		},
	)

	return sub, nil
}
