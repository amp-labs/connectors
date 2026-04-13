package calendly

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/require"
)

func TestBuildCalendlyEventList_eventTypes(t *testing.T) {
	t.Parallel()

	req := &SubscriptionRequest{}

	params := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"event_types": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
		},
	}

	got := buildCalendlyEventList(params, req)
	require.ElementsMatch(t, []string{
		"event_type.created",
		"event_type.updated",
		"event_type.deleted",
	}, got)
}

func TestSplitEventName(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"event_type", "created"}, splitEventName("event_type.created"))
	require.Equal(t, []string{"routing_form_submission", "created"}, splitEventName("routing_form_submission.created"))
}

func TestCalendlyHMACHex(t *testing.T) {
	t.Parallel()

	key := "test-signing-key"
	ts := "1234567890"
	body := []byte(`{"hello":"world"}`)

	expected := calendlyHMACHex(key, ts, body)

	// Deterministic: same inputs produce same hash.
	require.Equal(t, expected, calendlyHMACHex(key, ts, body))
	require.Len(t, expected, 64)
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
