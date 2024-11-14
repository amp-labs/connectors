package hubspot

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestWebhook(t *testing.T) {
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
