package jobber

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

const testWebhookPayload = `{
	"data": {
		"webHookEvent": {
			"topic": "CLIENT_CREATE",
			"appId": "3ef22a50-072d-430c-a78f-b7646657560b",
			"accountId": "Z2lkOi8vSm9iYmVyL0FjY291bnQvMjQ4NjkzNA==",
			"itemId": "Z2lkOi8vSm9iYmVyL0NsaWVudC8xNDUxODkzMjY=",
			"occurredAt": "2026-07-02T09:08:19Z"
		}
	}
}`

func parseTestEvent(t *testing.T, payload string) SubscriptionEvent {
	t.Helper()

	var evt SubscriptionEvent

	assert.NilError(t, json.Unmarshal([]byte(payload), &evt))

	return evt
}

func signBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	conn := &Connector{}
	secret := "test-client-secret"
	body := []byte(testWebhookPayload)

	t.Run("valid signature", func(t *testing.T) {
		t.Parallel()

		headers := http.Header{}
		headers.Set(WebhookSignatureHeader, signBody(secret, body))

		ok, err := conn.VerifyWebhookMessage(context.Background(),
			&common.WebhookRequest{Headers: headers, Body: body},
			&common.VerificationParams{Param: &JobberVerificationParams{Secret: secret}},
		)
		assert.NilError(t, err)
		assert.Assert(t, ok)
	})

	t.Run("wrong secret", func(t *testing.T) {
		t.Parallel()

		headers := http.Header{}
		headers.Set(WebhookSignatureHeader, signBody("wrong-secret", body))

		ok, err := conn.VerifyWebhookMessage(context.Background(),
			&common.WebhookRequest{Headers: headers, Body: body},
			&common.VerificationParams{Param: &JobberVerificationParams{Secret: secret}},
		)
		assert.Assert(t, errors.Is(err, ErrInvalidSignature))
		assert.Assert(t, !ok)
	})

	t.Run("missing header", func(t *testing.T) {
		t.Parallel()

		ok, err := conn.VerifyWebhookMessage(context.Background(),
			&common.WebhookRequest{Headers: http.Header{}, Body: body},
			&common.VerificationParams{Param: &JobberVerificationParams{Secret: secret}},
		)
		assert.Assert(t, errors.Is(err, ErrMissingSignature))
		assert.Assert(t, !ok)
	})

	t.Run("nil params", func(t *testing.T) {
		t.Parallel()

		ok, err := conn.VerifyWebhookMessage(context.Background(), nil, nil)
		assert.Assert(t, err != nil)
		assert.Assert(t, !ok)
	})
}

func TestSubscriptionEvent_Parsing(t *testing.T) {
	t.Parallel()

	evt := parseTestEvent(t, testWebhookPayload)

	eventType, err := evt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, eventType, common.SubscriptionEventTypeCreate)

	rawName, err := evt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, rawName, "CLIENT_CREATE")

	objectName, err := evt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, objectName, "clients")

	workspace, err := evt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, workspace, "Z2lkOi8vSm9iYmVyL0FjY291bnQvMjQ4NjkzNA==")

	recordID, err := evt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, recordID, "Z2lkOi8vSm9iYmVyL0NsaWVudC8xNDUxODkzMjY=")

	timestamp, err := evt.EventTimeStampNano()
	assert.NilError(t, err)

	expected, _ := time.Parse(time.RFC3339, "2026-07-02T09:08:19Z")
	assert.Equal(t, timestamp, expected.UnixNano())
}

func TestSubscriptionEvent_TopicMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		topic      string
		wantObject string
		wantType   common.SubscriptionEventType
	}{
		{"CLIENT_UPDATE", "clients", common.SubscriptionEventTypeUpdate},
		{"CLIENT_DESTROY", "clients", common.SubscriptionEventTypeDelete},
		{"PRODUCT_OR_SERVICE_CREATE", "products", common.SubscriptionEventTypeCreate},
		{"TIMESHEET_UPDATE", "timeSheetEntries", common.SubscriptionEventTypeUpdate},
		{"PAYOUT_CREATE", "payoutRecords", common.SubscriptionEventTypeCreate},
		{"QUOTE_SENT", "quotes", common.SubscriptionEventTypeOther},
		{"JOB_CLOSED", "jobs", common.SubscriptionEventTypeOther},
		{"VISIT_COMPLETE", "visits", common.SubscriptionEventTypeOther},
		// Topics without a connector object fall back to the raw topic.
		{"PAYMENT_CREATE", "PAYMENT_CREATE", common.SubscriptionEventTypeCreate},
		{"APP_DISCONNECT", "APP_DISCONNECT", common.SubscriptionEventTypeOther},
	}

	for _, tc := range cases {
		t.Run(tc.topic, func(t *testing.T) {
			t.Parallel()

			evt := parseTestEvent(t, testWebhookPayload)
			data := evt["data"].(map[string]any)           //nolint:forcetypeassert
			event := data["webHookEvent"].(map[string]any) //nolint:forcetypeassert
			event["topic"] = tc.topic

			objectName, err := evt.ObjectName()
			assert.NilError(t, err)
			assert.Equal(t, objectName, tc.wantObject)

			eventType, err := evt.EventType()
			assert.NilError(t, err)
			assert.Equal(t, eventType, tc.wantType)
		})
	}
}

func TestSubscriptionEvent_MisspelledOccuredAtFallback(t *testing.T) {
	t.Parallel()

	// Apps created before 2023-12-08 receive "occuredAt" instead of "occurredAt".
	payload := `{
		"data": {
			"webHookEvent": {
				"topic": "CLIENT_CREATE",
				"accountId": "MQ==",
				"itemId": "MQ==",
				"occuredAt": "2021-08-12T16:31:36-06:00"
			}
		}
	}`

	evt := parseTestEvent(t, payload)

	timestamp, err := evt.EventTimeStampNano()
	assert.NilError(t, err)

	expected, _ := time.Parse(time.RFC3339, "2021-08-12T16:31:36-06:00")
	assert.Equal(t, timestamp, expected.UnixNano())
}

func TestSubscriptionEvent_MissingFields(t *testing.T) {
	t.Parallel()

	evt := parseTestEvent(t, `{"data": {}}`)

	_, err := evt.RawEventName()
	assert.Assert(t, errors.Is(err, errMissingEventField))
}

func TestCollapsedSubscriptionEvent_SingleEvent(t *testing.T) {
	t.Parallel()

	var collapsed CollapsedSubscriptionEvent

	assert.NilError(t, json.Unmarshal([]byte(testWebhookPayload), &collapsed))

	events, err := collapsed.SubscriptionEventList()
	assert.NilError(t, err)
	assert.Equal(t, len(events), 1)

	rawName, err := events[0].RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, rawName, "CLIENT_CREATE")
}
