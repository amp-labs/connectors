package webhook

import (
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type Event map[string]any

var (
	_ common.SubscriptionEvent       = Event{}
	_ common.SubscriptionUpdateEvent = Event{}

	ErrMissingField = errors.New("missing field")
	ErrInvalidType  = errors.New("invalid type")
)

func (e Event) EventType() (common.SubscriptionEventType, error) {
	changeType, ok := e["changeType"].(string)
	if !ok {
		return "", ErrMissingField
	}

	switch strings.ToLower(changeType) {
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

func (e Event) RawEventName() (string, error) {
	changeType, ok := e["changeType"].(string)
	if !ok {
		return "", ErrMissingField
	}
	return changeType, nil
}

func (e Event) ObjectName() (string, error) {
	resource, ok := e["resource"].(string)
	if !ok {
		return "", ErrMissingField
	}
	// Resource is usually "Users/ID", "Messages/ID", etc.
	parts := strings.Split(resource, "/")
	if len(parts) > 0 {
		return parts[0], nil
	}
	return resource, nil
}

func (e Event) Workspace() (string, error) {
	// Microsoft doesn't really have a workspace in the same way, maybe tenantId?
	// But it's not always in the event.
	return "", nil
}

func (e Event) RecordId() (string, error) {
	resourceData, ok := e["resourceData"].(map[string]any)
	if ok {
		if id, ok := resourceData["id"].(string); ok {
			return id, nil
		}
	}

	// fallback to resource string
	resource, ok := e["resource"].(string)
	if ok {
		parts := strings.Split(resource, "/")
		if len(parts) > 1 {
			return parts[len(parts)-1], nil
		}
	}
	return "", ErrMissingField
}

func (e Event) EventTimeStampNano() (int64, error) {
	// Subscription expiration time is often present, but not event time.
	// However, some events might have it.
	return 0, nil
}

func (e Event) RawMap() (map[string]any, error) {
	return e, nil
}

func (e Event) PreLoadData(data *common.SubscriptionEventPreLoadData) error {
	return nil
}

func (e Event) UpdatedFields() ([]string, error) {
	// Microsoft doesn't usually provide a list of updated fields in the event itself
	// unless it's a "rich" notification with resource data.
	return nil, nil
}
