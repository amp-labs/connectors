// nolint:lll,tagliatelle
package salesforce

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

const (
	keyEventRecordID          = "recordId"
	keyEventRecordIdentifiers = "recordIds"
	keyEventChangeEventHeader = "ChangeEventHeader"
)

var (
	errChangeEventHeaderType = errors.New("key ChangeEventHeader is not of map[string]any type")
	errRecordIDsType         = errors.New("key recordIds is not of []any type")
)

func (*Connector) VerifyWebhookMessage(context.Context, *common.WebhookVerificationParameters) (bool, error) {
	return true, nil
}

var _ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}

// ChangeEvent represents data received from a subscription.
// A single event may contain multiple record identifiers and can be expanded into multiple SubscriptionEvent instances.
//
// Structure reference:
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_event_fields_header.htm.
type CollapsedSubscriptionEvent map[string]any

// ToRecordList splits bundled event into per record event.
// Every property is duplicated across SubscriptionEvent. RecordIds is spread as RecordId.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	eventHeaderMap, err := extractChangeEventHeader(e)
	if err != nil {
		return nil, err
	}

	recordIDsAny, err := eventHeaderMap.Get(keyEventRecordIdentifiers)
	if err != nil {
		return nil, err
	}

	recordIDs, ok := recordIDsAny.([]any)
	if !ok {
		return nil, errRecordIDsType
	}

	events := make([]common.SubscriptionEvent, len(recordIDs))

	for index, recordID := range recordIDs {
		event, err := goutils.Clone[map[string]any](e)
		if err != nil {
			return nil, err
		}

		subEvent := SubscriptionEvent(event)

		evt := common.SubscriptionEvent(subEvent)

		// Reach out to the nested object and remove record identifiers and attach record id.
		changeEventHeader, ok := event[keyEventChangeEventHeader].(map[string]any)
		if !ok {
			return nil, errChangeEventHeaderType
		}

		changeEventHeader[keyEventRecordID] = recordID
		delete(changeEventHeader, keyEventRecordIdentifiers)

		// Save changes.
		event[keyEventChangeEventHeader] = changeEventHeader
		events[index] = evt
	}

	return events, nil
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

// SubscriptionEvent holds event data.
//
// Record ID can have wildcard symbols:
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_field_conversion_single_event.htm.
type SubscriptionEvent map[string]any

func (s SubscriptionEvent) asMap() (common.StringMap, error) {
	return extractChangeEventHeader(s)
}

// EventType
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_event_fields_header.htm.
func (s SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	changeType, err := s.RawEventName()
	if err != nil {
		return "", err
	}

	switch changeType {
	case "CREATE":
		return common.SubscriptionEventTypeCreate, nil
	case "UPDATE":
		return common.SubscriptionEventTypeUpdate, nil
	case "DELETE":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (s SubscriptionEvent) RawEventName() (string, error) {
	registry, err := s.asMap()
	if err != nil {
		return "", err
	}

	return registry.GetString("changeType")
}

func (s SubscriptionEvent) ObjectName() (string, error) {
	registry, err := s.asMap()
	if err != nil {
		return "", err
	}

	// The API name of the standard or custom object that the change pertains to.
	// For example, Account or MyObject__c.
	return registry.GetString("entityName")
}

func (s SubscriptionEvent) Workspace() (string, error) {
	// Not applicable
	return "", nil
}

// RecordId
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_field_conversion_single_event.htm.
func (s SubscriptionEvent) RecordId() (string, error) {
	registry, err := s.asMap()
	if err != nil {
		return "", err
	}

	return registry.GetString(keyEventRecordID)
}

func (s SubscriptionEvent) EventTimeStampNano() (int64, error) {
	registry, err := s.asMap()
	if err != nil {
		return 0, err
	}

	// The date and time when the change occurred,
	// represented as the number of milliseconds since January 1, 1970 00:00:00 GMT.
	num, err := registry.GetNumber("commitTimestamp")
	if err != nil {
		return 0, err
	}

	return int64(num), nil
}

func extractChangeEventHeader(registry map[string]any) (common.StringMap, error) {
	eventMap := common.StringMap(registry)

	eventHeaderAny, err := eventMap.Get(keyEventChangeEventHeader)
	if err != nil {
		return nil, err
	}

	eventHeader, ok := eventHeaderAny.(map[string]any)
	if !ok {
		return nil, errChangeEventHeaderType
	}

	return eventHeader, nil
}
