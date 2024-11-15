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

	conn := &Connector{}

	objectName, err := conn.ExtractObjectNameFromWebhookMessage(correctMessage)
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

	_, err = conn.ExtractObjectNameFromWebhookMessage(unsupportedObjectMessage)
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

	_, err = conn.ExtractObjectNameFromWebhookMessage(emptyObjectMessage)
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

	conn := &Connector{}

	evtTypeCreate, err := conn.ExtractEventTypeFromWebhookMessage(createMessage)
	if err != nil {
		t.Fatalf("error extracting object name from webhook message: %s", err)
	}

	assert.Equal(t, evtTypeCreate, common.WebhookEventTypeCreate, "object type should be parsedCorrectly")

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

	evtTypeDelete, err := conn.ExtractEventTypeFromWebhookMessage(deleteMessage)
	if err != nil {
		t.Fatalf("error extracting object name from webhook message: %s", err)
	}

	assert.Equal(t, evtTypeDelete, common.WebhookEventTypeDelete, "object type should be parsedCorrectly")

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

	evtTypeUpdate, err := conn.ExtractEventTypeFromWebhookMessage(updateMessage)
	if err != nil {
		t.Fatalf("error extracting object name from webhook message: %s", err)
	}

	assert.Equal(t, evtTypeUpdate, common.WebhookEventTypeUpdate, "object type should be parsedCorrectly")

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

	_, err = conn.ExtractEventTypeFromWebhookMessage(unsupportedObjectMessage)
	assert.ErrorIs(t, err, errWebhookNotSupportedForObject)

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

	_, err = conn.ExtractEventTypeFromWebhookMessage(emptyObjectMessage)
	assert.ErrorIs(t, err, errWebhookNotSupportedForObject)
}
