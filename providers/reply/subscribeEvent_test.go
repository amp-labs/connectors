package reply

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

// envelope builds the event envelope every Reply delivery starts with.
func envelope(eventType, date string) map[string]any {
	return map[string]any{
		"date":            date,
		"id":              "4c1b9e74-5ac3-4b0f-9a60-2a37a9b4a2d1",
		"type":            eventType,
		"user_id":         float64(1234),
		"team_id":         float64(100),
		"subscription_id": float64(42),
	}
}

func TestSubscriptionEvent(t *testing.T) { //nolint:funlen
	t.Parallel()

	const eventDate = "2026-04-21T09:30:00.000Z"

	eventTime := time.Date(2026, time.April, 21, 9, 30, 0, 0, time.UTC).UnixNano()

	for _, tt := range []testconn.TestCaseSubscriptionEvent{
		{
			Name: "Contact event resolves the contact id",
			Input: CollapsedSubscriptionEvent{
				"event":          envelope("contact_opted_out", eventDate),
				"opt_out_date":   eventDate,
				"contact_fields": map[string]any{"id": float64(12345), "email": "test-prospect@example.com"},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "contact_opted_out",
					ObjectName:         "contacts",
					RecordId:           "12345",
					EventTimeStampNano: eventTime,
				},
			}},
		},
		{
			Name: "Email event resolves the contact id",
			Input: CollapsedSubscriptionEvent{
				"event":          envelope("email_replied", eventDate),
				"sent_email_id":  float64(999),
				"contact_fields": map[string]any{"id": float64(761575999)},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "email_replied",
					ObjectName:         "contacts",
					RecordId:           "761575999",
					EventTimeStampNano: eventTime,
				},
			}},
		},
		{
			Name: "Email account event resolves the email account id",
			Input: CollapsedSubscriptionEvent{
				"event":               envelope("email_account_error", eventDate),
				"email_account_id":    float64(1),
				"email_account_error": "Sending: SMTP authentication failed",
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "email_account_error",
					ObjectName:         "email-accounts",
					RecordId:           "1",
					EventTimeStampNano: eventTime,
				},
			}},
		},
		{
			Name: "LinkedIn account alert resolves the LinkedIn account id",
			Input: CollapsedSubscriptionEvent{
				"event":       envelope("linkedin_account_alerts", eventDate),
				"linkedin_id": float64(50),
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "linkedin_account_alerts",
					ObjectName:         "linkedin-accounts",
					RecordId:           "50",
					EventTimeStampNano: eventTime,
				},
			}},
		},
		{
			Name: "Autopilot event resolves the sequence id",
			Input: CollapsedSubscriptionEvent{
				"event":           envelope("autopilot_stopped", eventDate),
				"sequence_fields": map[string]any{"id": float64(200), "name": "Test Sequence"},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "autopilot_stopped",
					ObjectName:         "sequences",
					RecordId:           "200",
					EventTimeStampNano: eventTime,
				},
			}},
		},
		{
			Name: "Event time keeps the sub-second precision Reply sends",
			Input: CollapsedSubscriptionEvent{
				"event":          envelope("contact_finished", "2026-10-30T14:05:55.4146886Z"),
				"contact_fields": map[string]any{"id": float64(12345)},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "contact_finished",
					ObjectName:         "contacts",
					RecordId:           "12345",
					EventTimeStampNano: time.Date(2026, time.October, 30, 14, 5, 55, 414688600, time.UTC).UnixNano(),
				},
			}},
		},
		{
			Name: "Payload without an event envelope is rejected",
			Input: SubscriptionEvent{
				"contact_fields": map[string]any{"id": float64(12345)},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType: common.SubscriptionEventTypeOther,
				},
				Err: testconn.SubscriptionEventExpectedErr{
					EventType:          testutils.StringError("reply event: missing event.type"),
					RawEventName:       testutils.StringError("reply event: missing event.type"),
					ObjectName:         testutils.StringError("reply event: missing event.type"),
					RecordId:           testutils.StringError("reply event: missing event.type"),
					EventTimeStampNano: testutils.StringError("reply event: missing event.date"),
				},
			}},
		},
		{
			Name: "Event type Reply does not offer is rejected",
			Input: SubscriptionEvent{
				"event":          envelope("email_clicked", eventDate),
				"contact_fields": map[string]any{"id": float64(12345)},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "email_clicked",
					EventTimeStampNano: eventTime,
				},
				Err: testconn.SubscriptionEventExpectedErr{
					EventType:  testutils.StringError("reply event: event type is not supported: email_clicked"),
					ObjectName: testutils.StringError("reply event: event type is not supported: email_clicked"),
					RecordId:   testutils.StringError("reply event: event type is not supported: email_clicked"),
				},
			}},
		},
		{
			Name: "Contact event without a contact id is rejected",
			Input: SubscriptionEvent{
				"event": envelope("email_sent", eventDate),
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeOther,
					RawEventName:       "email_sent",
					ObjectName:         "contacts",
					EventTimeStampNano: eventTime,
				},
				Err: testconn.SubscriptionEventExpectedErr{
					RecordId: testutils.StringError("reply event: missing record id in payload"),
				},
			}},
		},
		{
			Name: "Unparsable event date is rejected",
			Input: SubscriptionEvent{
				"event":          envelope("contact_called", "yesterday"),
				"contact_fields": map[string]any{"id": float64(12345)},
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    common.SubscriptionEventTypeOther,
					RawEventName: "contact_called",
					ObjectName:   "contacts",
					RecordId:     "12345",
				},
				Err: testconn.SubscriptionEventExpectedErr{
					EventTimeStampNano: testutils.StringError("reply event: unparsable event.date"),
				},
			}},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t)
		})
	}
}

// TestCapturedDeliveries parses deliveries as Reply sent them: each file is the body of a real
// POST /v3/webhooks/{id}/test delivery captured at the receiving end (only the user id is replaced).
func TestCapturedDeliveries(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		eventType string
		object    string
		recordID  string
	}{
		{"contact_opted_out", "contacts", "12345"},
		{"email_replied", "contacts", "12345"},
		{"email_account_error", "email-accounts", "1"},
		{"linkedin_account_alerts", "linkedin-accounts", "50"},
		{"autopilot_stopped", "sequences", "200"},
	} {
		t.Run(tt.eventType, func(t *testing.T) {
			t.Parallel()

			var delivery map[string]any

			err := json.Unmarshal(testutils.DataFromFile(t, "subscribe/"+tt.eventType+".json"), &delivery)
			assert.NilError(t, err)

			occurred, err := time.Parse(time.RFC3339Nano, delivery["event"].(map[string]any)["date"].(string))
			assert.NilError(t, err)

			testconn.TestCaseSubscriptionEvent{
				Name:  tt.eventType,
				Input: CollapsedSubscriptionEvent(delivery),
				Expected: []testconn.SubscriptionEventExpected{{
					Data: testconn.SubscriptionEventExpectedData{
						EventType:          common.SubscriptionEventTypeOther,
						RawEventName:       tt.eventType,
						ObjectName:         tt.object,
						RecordId:           tt.recordID,
						EventTimeStampNano: occurred.UnixNano(),
					},
				}},
			}.Run(t)
		})
	}
}

// documentedEvents is the catalog of event types Reply lists on
// https://docs.reply.io/webhook-events and returns from GET /v3/webhooks/events.
//
//nolint:gochecknoglobals
var documentedEvents = []string{
	"email_sent", "email_opened", "email_link_clicked", "email_replied",
	"reply_categorized", "email_bounced", "email_auto_reply",
	"email_account_connection_lost", "email_account_error",
	"linkedin_connection_request_sent", "linkedin_connection_request_accepted",
	"linkedin_message_sent", "linkedin_message_replied", "linkedin_reply_categorized",
	"linkedin_account_alerts", "linkedin_comment_posted", "linkedin_comment_skipped",
	"contact_called", "contact_replied", "contact_finished", "contact_opted_out",
	"autopilot_stopped",
}

func TestEveryDocumentedEventIsRouted(t *testing.T) {
	t.Parallel()

	assert.Equal(t, len(eventRoutes), len(documentedEvents))

	for _, eventType := range documentedEvents {
		t.Run(eventType, func(t *testing.T) {
			t.Parallel()

			route, ok := eventRoutes[eventType]
			assert.Assert(t, ok, "event %q has no route", eventType)
			assert.Assert(t, slices.Contains(
				[]string{"contacts", "email-accounts", "linkedin-accounts", "sequences"}, route.object,
			))
			assert.Assert(t, recordObjects.Has(route.object),
				"object %q has no get-by-id endpoint, so its events could not be enriched", route.object)
			assert.Assert(t, len(route.idPath) > 0)
		})
	}
}

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	// Reply sends no signature, so every delivery is accepted. The verifier is the zero value
	// connector the subscribe registry uses, which has no authenticated client.
	for _, tt := range []testconn.TestCaseVerifyWebhookMessage{
		{
			Name: "Delivery without any signature header is accepted",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
					Body:    []byte(`{"event":{"type":"contact_opted_out"}}`),
				},
				Params: &common.VerificationParams{},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
		{
			Name: "Missing request and params are accepted",
			Input: testconn.WebhookMessageVerificationParams{
				Request: nil,
				Params:  nil,
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWebhookMessageVerifier, error) {
				return &Connector{}, nil
			})
		})
	}
}

// The subscribe registry registers a zero value Connector as the webhook verifier and the
// server calls Provider and String on it, which must not dereference the missing base.
func TestZeroValueConnectorIdentity(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	assert.Equal(t, conn.Provider(), providers.Reply)
	assert.Equal(t, conn.String(), providers.Reply+".Connector")

	var nilConn *Connector

	assert.Equal(t, nilConn.String(), providers.Reply+".Connector")
}
