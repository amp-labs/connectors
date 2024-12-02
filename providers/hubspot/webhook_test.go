package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestExtractObjectNameFromWebhookMessage(t *testing.T) {
	t.Parallel()

	correctMessage := &WebhookMessage{
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

	objectName, err := correctMessage.ObjectName()
	if err != nil {
		t.Fatalf("error extracting object name from webhook message: %s", err)
	}

	assert.Equal(t, objectName, "contact", "object name should be parsedCorrectly")

	unsupportedObjectMessage := &WebhookMessage{
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

	_, err = unsupportedObjectMessage.ObjectName()
	assert.ErrorContains(t, err, "webhook is not supported for the object 'someObject'")

	emptyObjectMessage := &WebhookMessage{
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

	_, err = emptyObjectMessage.ObjectName()
	assert.ErrorContains(t, err, "webhook is not supported for the object ''")
}

//nolint:funlen
func TestExtractEventTypeFromWebhookMessage(t *testing.T) {
	t.Parallel()

	createMessage := &WebhookMessage{
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

	evtTypeCreate, err := createMessage.EventType()
	if err != nil {
		t.Fatalf("error extracting object name from webhook message: %s", err)
	}

	assert.Equal(t, evtTypeCreate, common.WebhookEventTypeCreate, "event type should be parsed Correctly")

	deleteMessage := &WebhookMessage{
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
		t.Fatalf("error extracting eventTye from webhook message: %s", err)
	}

	assert.Equal(t, evtTypeDelete, common.WebhookEventTypeDelete, "event type should be parsed correctly")

	updateMessage := &WebhookMessage{
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
		t.Fatalf("error extracting eventTye from webhook message: %s", err)
	}

	assert.Equal(t, evtTypeUpdate, common.WebhookEventTypeUpdate, "event type should be parsed correctly")

	emptyObjectMessage := &WebhookMessage{
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

	_, err = emptyObjectMessage.EventType()
	assert.ErrorIs(t, err, errUnexpectedWebhookEventType)
}
