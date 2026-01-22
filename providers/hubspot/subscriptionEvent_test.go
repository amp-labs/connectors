package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestExtractObjectNameFromSubscriptionEvent(t *testing.T) {
	t.Parallel()

	validEvent := SubscriptionEvent{
		"subscriptionType": "contact.creation",
	}

	objectName, err := validEvent.ObjectName()
	if err != nil {
		t.Fatalf("error extracting object name from subscription event: %s", err)
	}

	assert.Equal(t, objectName, "contact", "object name should be parsedCorrectly")

	unsupportedEvent := SubscriptionEvent{
		"subscriptionType": "someObject.creation",
	}

	_, err = unsupportedEvent.ObjectName()
	assert.ErrorContains(t, err, "subscription is not supported for the object 'someObject'")

	emptyObjectEvent := &SubscriptionEvent{
		"subscriptionType": "",
	}

	_, err = emptyObjectEvent.ObjectName()
	assert.ErrorContains(t, err, "subscription is not supported for the object ''")

	withObjectTypeId := SubscriptionEvent{
		"objectTypeId": "0-1",
	}

	objectName, err = withObjectTypeId.ObjectName()
	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, objectName, "contact", "object name should be parsed correctly")
}

//nolint:funlen
func TestExtractEventTypeFromSubscriptionEvent(t *testing.T) {
	t.Parallel()

	createEvent := SubscriptionEvent{
		"subscriptionType": "contact.creation",
	}

	evtTypeCreate, err := createEvent.EventType()
	if err != nil {
		t.Fatalf("error extracting object name from subscription  event: %s", err)
	}

	assert.Equal(t, evtTypeCreate, common.SubscriptionEventTypeCreate, "event type should be parsed Correctly")

	deleteMessage := SubscriptionEvent{
		"subscriptionType": "contact.deletion",
	}

	evtTypeDelete, err := deleteMessage.EventType()
	if err != nil {
		t.Fatalf("error extracting eventType from subscription event: %s", err)
	}

	assert.Equal(t, evtTypeDelete, common.SubscriptionEventTypeDelete, "event type should be parsed correctly")

	updateMessage := SubscriptionEvent{
		"subscriptionType": "contact.propertyChange",
	}

	evtTypeUpdate, err := updateMessage.EventType()
	if err != nil {
		t.Fatalf("error extracting eventType from subscription event: %s", err)
	}

	assert.Equal(t, evtTypeUpdate, common.SubscriptionEventTypeUpdate, "event type should be parsed correctly")

	emptyObjectEvent := SubscriptionEvent{
		"subscriptionType": "",
	}

	_, err = emptyObjectEvent.EventType()
	assert.ErrorIs(
		t,
		err,
		errUnexpectedSubscriptionEventType,
		"error should be of type errUnexpectedSubscriptionEventType",
	)
}
