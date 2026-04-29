package webhook

import (
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
)

// EventCollection is a change notification sent by Microsoft Graph to the webhook.
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#change-notification-example
type EventCollection map[string]any

// Event is a singular notification message within EventCollection.
type Event map[string]any

var (
	_ common.CollapsedSubscriptionEvent = EventCollection{}
	_ common.SubscriptionEvent          = Event{}
	_ common.SubscriptionUpdateEvent    = Event{}

	ErrMissingField = errors.New("missing field")
)

func (c EventCollection) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	value, ok := c["value"]
	if !ok {
		return nil, fmt.Errorf("%w: missing key 'value'", common.ErrSubscriptionEventList)
	}

	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: 'value' is not []any type", common.ErrSubscriptionEventList)
	}

	events := make([]common.SubscriptionEvent, len(list))
	for index, item := range list {
		if json, ok := item.(map[string]any); !ok {
			return nil, fmt.Errorf(
				"%w: 'value[%v]' is not map[string]any", common.ErrSubscriptionEventList, index,
			)
		} else {
			events[index] = Event(json)
		}
	}

	return events, nil
}

func (c EventCollection) RawMap() (map[string]any, error) {
	return maps.Clone(c), nil
}

func (e Event) EventType() (common.SubscriptionEventType, error) {
	changeTypeStr, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	changeType := ChangeType(changeTypeStr)

	list := changeType.EventTypes()
	if len(list) == 0 {
		// There should be just one type in the event response.
		// However, when creating a subscription multiple types can be supplied,
		// hence the list nature of changeType property.
		return common.SubscriptionEventTypeOther, nil
	}

	return list[0], nil
}

func (e Event) RawEventName() (string, error) {
	changeTypeStr, ok := e["changeType"].(string)
	if !ok {
		return "", fmt.Errorf("%w: 'changeType'", ErrMissingField)
	}

	return changeTypeStr, nil
}

func (e Event) ObjectName() (string, error) {
	objectName, ok := e["clientState"].(string) // TODO before converting must first get the data first.
	if !ok {
		return "", fmt.Errorf("%w: 'clientState'", ErrMissingField)
	}

	return objectName, nil
}

func (e Event) Workspace() (string, error) {
	return "", nil
}

func (e Event) RecordId() (string, error) {
	resourceData, ok := e["resourceData"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: 'resourceData'", ErrMissingField)
	}

	identifier, ok := resourceData["id"].(string)
	if !ok {
		return "", fmt.Errorf("%w: 'id'", ErrMissingField)
	}

	return identifier, nil
}

func (e Event) EventTimeStampNano() (int64, error) {
	return 0, nil
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
