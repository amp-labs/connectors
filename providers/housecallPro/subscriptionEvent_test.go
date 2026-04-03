package housecallpro

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

func testWebhookHMACHex(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(body)))

	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		body      []byte
		timestamp string
		secret    string
		signature string
		wantOK    bool
		wantErr   bool
	}{
		{
			name:      "valid signature from fixture",
			body:      testutils.DataFromFile(t, "webhook-job-created.json"),
			timestamp: "1775143244",
			secret:    "test-webhook-hmac-secret",
			wantOK:    true,
			wantErr:   false,
		},
		{
			name:      "invalid signature",
			body:      []byte(`{"event":"job.created"}`),
			timestamp: "1775143244",
			secret:    "s",
			signature: "deadbeef",
			wantOK:    false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			signature := tt.signature
			if signature == "" {
				signature = testWebhookHMACHex(tt.secret, tt.timestamp, tt.body)
			}

			h := http.Header{}
			h.Set(housecallAPITimestampHeader, tt.timestamp)
			h.Set(housecallAPISignatureHeader, signature)

			ok, err := (&Connector{}).VerifyWebhookMessage(t.Context(), &common.WebhookRequest{
				Headers: h,
				Body:    tt.body,
			}, &common.VerificationParams{
				Param: &HousecallProVerificationParams{Secret: tt.secret},
			})

			assert.Equal(t, ok, tt.wantOK)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestCollapsedSubscriptionEvent_Interface(t *testing.T) {
	t.Parallel()

	tests := []fixtureExpectation{
		{name: "created", fixture: "webhook-job-created.json"},
		{name: "updated", fixture: "webhook-estimate-updated.json"},
		{name: "invoice refund succeeded", fixture: "webhook-invoice-refund-succeeded.json"},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evt := loadWebhookEventFixture(t, tt.fixture)

			rawMap, err := evt.RawMap()
			assert.NilError(t, err)
			assert.Assert(t, rawMap != nil)

			events, err := evt.SubscriptionEventList()
			assert.NilError(t, err)
			assert.Equal(t, len(events), 1)
		})
	}
}

func TestSubscriptionEvent_Interface(t *testing.T) {
	t.Parallel()

	tests := []fixtureExpectation{
		{
			name:    "created",
			fixture: "webhook-job-created.json",
			want: eventAssertions{
				eventType: common.SubscriptionEventTypeCreate,
				rawName:   "job.created",
				object:    "jobs",
				recordID:  "job_ac6f3efd11c14a5aa93e9fc0ab5354ab",
				payload:   "job",
			},
		},
		{
			name:    "updated",
			fixture: "webhook-estimate-updated.json",
			want: eventAssertions{
				eventType: common.SubscriptionEventTypeUpdate,
				rawName:   "estimate.updated",
				object:    "estimates",
				recordID:  "csr_fb441cf2bb9445019530fabbda25555e",
				payload:   "estimate",
			},
		},
		{
			name:    "invoice refund succeeded",
			fixture: "webhook-invoice-refund-succeeded.json",
			want: eventAssertions{
				eventType: common.SubscriptionEventTypeOther,
				rawName:   "invoice.refund.succeeded",
				object:    "invoices",
				recordID:  "invoice_5f3de0f1d9e1483f9a4e4be0c7a44f0b",
				payload:   "invoice",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSubscriptionEvent(t, tt.fixture, tt.want)
		})
	}
}

func TestSubscriptionEvent_UnsupportedJobAppointmentPrefix(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"event":             "job.appointment.appointment_discarded",
		"event_occurred_at": "2026-04-03T14:20:38Z",
		"company_id":        "7141dca7-882d-427b-a9c0-0ba0d74c85cf",
		"job": map[string]any{
			"id": "job_ac6f3efd11c14a5aa93e9fc0ab5354ab",
		},
	}

	_, err := evt.ObjectName()
	assert.Assert(t, err != nil)

	_, err = evt.RecordId()
	assert.Assert(t, err != nil)
}

func TestSubscriptionEvent_ObjectNameRequiresPayload(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"event":             "job.created",
		"event_occurred_at": "2026-04-03T14:20:38Z",
		"company_id":        "7141dca7-882d-427b-a9c0-0ba0d74c85cf",
	}

	_, err := evt.ObjectName()
	assert.Assert(t, err != nil)
	assert.Assert(t, errors.Is(err, errMalformedWebhookEvent))

	_, err = evt.RecordId()
	assert.Assert(t, err != nil)
}

func TestSubscriptionEvent_ParsingWeirdSamples(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		event             string
		payloadKey        string
		payloadID         string
		wantEventType     common.SubscriptionEventType
		wantObjectName    string
		expectObjectError bool
	}{
		{
			name:           "invoice payment failed",
			event:          "invoice.payment.failed",
			payloadKey:     "invoice",
			payloadID:      "invoice_1a2b3c4d5e",
			wantEventType:  common.SubscriptionEventTypeOther,
			wantObjectName: "invoices",
		},
		{
			name:           "customer membership renewed",
			event:          "customer.membership.renewed",
			payloadKey:     "customer",
			payloadID:      "cus_9985725dd4824c56b2a1aef45986d1cc",
			wantEventType:  common.SubscriptionEventTypeOther,
			wantObjectName: "customers",
		},
		{
			name:              "malformed event name",
			event:             "invoice",
			payloadKey:        "invoice",
			payloadID:         "invoice_9f8e7d6c5b",
			wantEventType:     common.SubscriptionEventTypeOther,
			expectObjectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evt := SubscriptionEvent{
				"event":             tt.event,
				"event_occurred_at": "2026-04-03T15:20:38Z",
				"company_id":        "7141dca7-882d-427b-a9c0-0ba0d74c85cf",
				tt.payloadKey: map[string]any{
					"id": tt.payloadID,
				},
			}

			et, err := evt.EventType()
			assert.NilError(t, err)
			assert.Equal(t, et, tt.wantEventType)

			if tt.expectObjectError {
				_, err = evt.ObjectName()
				assert.Assert(t, err != nil)
				_, err = evt.RecordId()
				assert.Assert(t, err != nil)
			} else {
				obj, err := evt.ObjectName()
				assert.NilError(t, err)
				assert.Equal(t, obj, tt.wantObjectName)

				id, err := evt.RecordId()
				assert.NilError(t, err)
				assert.Equal(t, id, tt.payloadID)
			}
		})
	}
}

type fixtureExpectation struct {
	name    string
	fixture string
	want    eventAssertions
}

type eventAssertions struct {
	eventType common.SubscriptionEventType
	rawName   string
	object    string
	recordID  string
	payload   string
}

func assertSubscriptionEvent(t *testing.T, fixture string, want eventAssertions) {
	t.Helper()

	evt := loadWebhookEventFixture(t, fixture)

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err)
	assert.Equal(t, len(events), 1)

	subEvt := events[0]

	eventType, err := subEvt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, eventType, want.eventType)

	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, rawEventName, want.rawName)

	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, objectName, want.object)

	recordID, err := subEvt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, recordID, want.recordID)

	workspace, err := subEvt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, workspace, "")

	timestamp, err := subEvt.EventTimeStampNano()
	assert.NilError(t, err)
	assert.Assert(t, timestamp > 0)

	evRawMap, err := subEvt.RawMap()
	assert.NilError(t, err)
	_, ok := evRawMap[want.payload]
	assert.Assert(t, ok)
}

func loadWebhookEventFixture(t *testing.T, filename string) CollapsedSubscriptionEvent {
	t.Helper()

	raw := testutils.DataFromFile(t, filename)
	if len(strings.TrimSpace(string(raw))) == 0 {
		t.Skipf("webhook fixture %q is empty", filename)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", filename, err)
	}

	return CollapsedSubscriptionEvent(payload)
}
