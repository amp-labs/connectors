package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestExtractObjectNameFromSubscriptionEvent(t *testing.T) {
	t.Parallel()

	validEvent := &SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "contact.creation",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	objectName, err := validEvent.ObjectName()
	if err != nil {
		t.Fatalf("error extracting object name from subscription event: %s", err)
	}

	assert.Equal(t, objectName, "contact", "object name should be parsedCorrectly")

	unsupportedEvent := &SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "someObject.creation",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	_, err = unsupportedEvent.ObjectName()
	assert.ErrorContains(t, err, "subscription is not supported for the object 'someObject'")

	emptyObjectEvent := &SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	_, err = emptyObjectEvent.ObjectName()
	assert.ErrorContains(t, err, "subscription is not supported for the object ''")
}

//nolint:funlen
func TestExtractEventTypeFromSubscriptionEvent(t *testing.T) {
	t.Parallel()

	createEvent := SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "contact.creation",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	evtTypeCreate, err := createEvent.EventType()
	if err != nil {
		t.Fatalf("error extracting object name from subscription  event: %s", err)
	}

	assert.Equal(t, evtTypeCreate, common.SubscriptionEventTypeCreate, "event type should be parsed Correctly")

	deleteMessage := &SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "contact.deletion",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	evtTypeDelete, err := deleteMessage.EventType()
	if err != nil {
		t.Fatalf("error extracting eventTye from subscription event: %s", err)
	}

	assert.Equal(t, evtTypeDelete, common.SubscriptionEventTypeDelete, "event type should be parsed correctly")

	updateMessage := &SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "contact.propertyChange",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	evtTypeUpdate, err := updateMessage.EventType()
	if err != nil {
		t.Fatalf("error extracting eventTye from subscription event: %s", err)
	}

	assert.Equal(t, evtTypeUpdate, common.SubscriptionEventTypeUpdate, "event type should be parsed correctly")

	emptyObjectEvent := &SubscriptionEvent{
		AppId:            1,
		EventId:          1,
		SubscriptionId:   1,
		PortalId:         1,
		OccurredAt:       1,
		SubscriptionType: "",
		AttemptNumber:    1,
		ObjectId:         1,
		ChangeSource:     "CRM",
		PropertyName:     "message",
		PropertyValue:    "sample-value",
	}

	_, err = emptyObjectEvent.EventType()
	assert.ErrorIs(
		t,
		err,
		errUnexpectedSubscriptionEventType,
		"error should be of type errUnexpectedSubscriptionEventType",
	)
}
