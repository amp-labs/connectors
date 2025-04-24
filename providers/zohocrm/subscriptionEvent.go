package zohocrm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
)

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}
)

// SubscriptionEvent represents a webhook event from Zoho CRM
type SubscriptionEvent map[string]any

// VerifyWebhookMessage verifies the signature of a webhook message from Zoho CRM
func (*Connector) VerifyWebhookMessage(
	_ context.Context, params *common.WebhookVerificationParameters,
) (bool, error) {
	return true, nil
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

// EventType returns the type of event (create, update, delete)
func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	operation, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	switch operation {
	case "create", "insert":
		return common.SubscriptionEventTypeCreate, nil
	case "edit", "update":
		return common.SubscriptionEventTypeUpdate, nil
	case "delete":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

// RawEventName returns the raw event name from the zoho crm
func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	operation, err := m.GetString("operation")
	if err != nil {
		return "", err
	}

	return operation, nil
}

// ObjectName returns the name of the object that triggered the event
func (evt SubscriptionEvent) ObjectName() (string, error) {
	m := evt.asMap()

	module, err := m.GetString("module")
	if err != nil {
		return "", err
	}

	return module, nil
}

// Workspace returns the workspace ID
func (evt SubscriptionEvent) Workspace() (string, error) {
	m := evt.asMap()

	// In Zoho CRM, the channel_id can be used as the workspace identifier
	channelID, err := m.GetString("channel_id")
	if err != nil {
		return "", err
	}

	return channelID, nil
}

// RecordId returns the ID of the record that triggered the event
func (evt SubscriptionEvent) RecordId() (string, error) {
	m := evt.asMap()

	IdsAny, err := m.Get("id")
	if err != nil {
		return "", fmt.Errorf("errror getting record id: %w", err) //nolint:err113
	}

	//convert it to array
	ids, ok := IdsAny.([]any)
	if !ok || len(ids) == 0 {
		return "", errors.New("invalid or empty ids array") //nolint:err113
	}

	// Get the first ID
	id, ok := ids[0].(string)
	if !ok {
		return "", errors.New("invalid record id format") //nolint:err113
	}

	return id, nil
}

// EventTimeStampNano returns the timestamp of the event in nanoseconds
func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	m := evt.asMap()

	serverTime, err := m.AsInt("server_time")
	if err != nil {
		return 0, err
	}

	return time.UnixMilli(serverTime).UnixNano(), nil
}

// UpdatedFields returns the fields that were updated in the event
func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	m := evt.asMap()

	// Get the affected_fields array
	affectedFieldsAny, err := m.Get("affected_fields")
	if err != nil {
		return nil, err
	}

	// nolint:varnamelen
	affectedFields, ok := affectedFieldsAny.([]any)
	if !ok || len(affectedFields) == 0 {
		return nil, errors.New("invalid or empty affected_fields array")
	}

	// Get the first element which should be a map
	firstElement, ok := affectedFields[0].(map[string]any)
	if !ok {
		return nil, errInvalidField
	}

	// Get the record ID
	recordID, err := evt.RecordId()
	if err != nil {
		return nil, err
	}

	// Get the fields for this record
	fieldsAny, ok := firstElement[recordID]
	if !ok {
		return nil, fmt.Errorf("no fields found for record ID %s", recordID) //nolint:err113
	}

	// Convert to array of strings
	fieldsArray, ok := fieldsAny.([]any)
	if !ok {
		return nil, errInvalidField
	}

	fields := make([]string, 0, len(fieldsArray))

	for _, fieldAny := range fieldsArray {
		field, ok := fieldAny.(string)
		if !ok {
			return nil, errInvalidField
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// asMap returns the event as a StringMap
func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
}
