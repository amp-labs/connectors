package calendly

import (
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionEvent_eventType_nestedPayload(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"event": "event_type.updated",
		"payload": map[string]any{
			"event_type": "https://api.calendly.com/event_types/et-1",
			"updated_at": "2019-08-24T14:15:22.123456Z",
		},
	}

	raw, err := evt.RawEventName()
	require.NoError(t, err)
	require.Equal(t, "event_type.updated", raw)

	et, err := evt.EventType()
	require.NoError(t, err)
	require.Equal(t, common.SubscriptionEventTypeUpdate, et)

	obj, err := evt.ObjectName()
	require.NoError(t, err)
	require.Equal(t, "event_types", obj)

	id, err := evt.RecordId()
	require.NoError(t, err)
	require.Equal(t, "https://api.calendly.com/event_types/et-1", id)

	ts, err := evt.EventTimeStampNano()
	require.NoError(t, err)
	expected, err := time.Parse(time.RFC3339, "2019-08-24T14:15:22.123456Z")
	require.NoError(t, err)
	require.Equal(t, expected.UnixNano(), ts)

	fields, err := evt.UpdatedFields()
	require.NoError(t, err)
	require.Nil(t, fields)

	m, err := evt.RawMap()
	require.NoError(t, err)
	require.Equal(t, "event_type.updated", m["event"])
}

func TestSubscriptionEvent_eventType_flatBody(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"event":        "event_type.created",
		"organization": "https://api.calendly.com/organizations/org-1",
		"event_type":   "https://api.calendly.com/event_types/et-2",
		"created_at":   "2020-01-01T00:00:00Z",
	}

	org, err := evt.Workspace()
	require.NoError(t, err)
	require.Equal(t, "https://api.calendly.com/organizations/org-1", org)

	id, err := evt.RecordId()
	require.NoError(t, err)
	require.Equal(t, "https://api.calendly.com/event_types/et-2", id)

	et, err := evt.EventType()
	require.NoError(t, err)
	require.Equal(t, common.SubscriptionEventTypeCreate, et)
}

func TestSubscriptionEvent_inviteeCreated_scheduledEventURI(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"event": "invitee.created",
		"payload": map[string]any{
			"uri": "https://api.calendly.com/scheduled_events/guest-1",
			"scheduled_event": map[string]any{
				"uri": "https://api.calendly.com/scheduled_events/se-1",
				"organization": "https://api.calendly.com/organizations/org-9",
			},
		},
	}

	obj, err := evt.ObjectName()
	require.NoError(t, err)
	require.Equal(t, "scheduled_events", obj)

	ws, err := evt.Workspace()
	require.NoError(t, err)
	require.Equal(t, "https://api.calendly.com/organizations/org-9", ws)

	id, err := evt.RecordId()
	require.NoError(t, err)
	require.Equal(t, "https://api.calendly.com/scheduled_events/guest-1", id)
}

func TestSubscriptionEvent_ObjectName_unsupportedFamily(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{"event": "unknown.thing"}

	_, err := evt.ObjectName()
	require.ErrorContains(t, err, "unsupported webhook event family")
}

func TestSubscriptionEvent_RawEventName_missing(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{"payload": map[string]any{}}

	_, err := evt.RawEventName()
	require.Error(t, err)
}

func TestCalendlyHMACHex(t *testing.T) {
	t.Parallel()

	key := "test-signing-key"
	ts := "1234567890"
	body := []byte(`{"hello":"world"}`)

	expected := calendlyHMACHex(key, ts, body)
	require.Equal(t, expected, calendlyHMACHex(key, ts, body))
	require.Len(t, expected, 64)
}

func TestParseCalendlySignatureHeader(t *testing.T) {
	t.Parallel()

	ts, v1, err := parseCalendlySignatureHeader("t=99,v1=deadbeef")
	require.NoError(t, err)
	require.Equal(t, "99", ts)
	require.Equal(t, "deadbeef", v1)

	_, _, err = parseCalendlySignatureHeader("v1=only")
	require.Error(t, err)
}

func TestVerifyWebhookMessageSignature(t *testing.T) {
	t.Parallel()

	key := "test-signing-key"
	ts := "1234567890"
	body := []byte(`{"event":"event_type.created"}`)

	expectedHex := calendlyHMACHex(key, ts, body)
	header := "t=" + ts + ",v1=" + expectedHex

	ok, err := (&Connector{}).VerifyWebhookMessage(t.Context(), &common.WebhookRequest{
		Headers: map[string][]string{
			headerCalendlyWebhookSignature: {header},
		},
		Body: body,
	}, &common.VerificationParams{
		Param: &CalendlyVerificationParams{SigningKey: key},
	})
	require.NoError(t, err)
	require.True(t, ok)
}

func TestVerifyWebhookMessage_invalidParams(t *testing.T) {
	t.Parallel()

	ok, err := (&Connector{}).VerifyWebhookMessage(t.Context(), &common.WebhookRequest{
		Headers: map[string][]string{},
		Body:    []byte(`{}`),
	}, &common.VerificationParams{Param: "wrong-type"})
	require.False(t, ok)
	require.Error(t, err)
}

func TestVerifyWebhookMessage_missingHeader(t *testing.T) {
	t.Parallel()

	ok, err := (&Connector{}).VerifyWebhookMessage(t.Context(), &common.WebhookRequest{
		Headers: map[string][]string{},
		Body:    []byte(`{}`),
	}, &common.VerificationParams{
		Param: &CalendlyVerificationParams{SigningKey: "k"},
	})
	require.False(t, ok)
	require.Error(t, err)
}

func TestSecureCompareHex_mismatch(t *testing.T) {
	t.Parallel()

	ok, err := secureCompareHex("ab", "cd")
	require.NoError(t, err)
	require.False(t, ok)
}
