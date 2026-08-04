package zoho

import (
	"bytes"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

// crmWebhookBody is a Zoho CRM notification payload. server_time is a bare JSON
// number (epoch millis).
const crmWebhookBody = `{
	"server_time": 1750102639787,
	"affected_values": [
		{"record_id": "6756839000000575405", "values": {"Company": "Rangoni Of Test"}}
	],
	"module": "Leads",
	"ids": ["6756839000000575405"],
	"affected_fields": [{"6756839000000575405": ["Company"]}],
	"operation": "update",
	"channel_id": "1105420521999070702",
	"token": "c3504777-db15-4332-8286-478a1b5006bc"
}`

// TestEventTimeStampNanoDecodeModes verifies the CRM event timestamp is read
// correctly whether the caller decoded the webhook body with a plain
// json.Unmarshal (server_time -> float64) or with json.Decoder.UseNumber
// (server_time -> json.Number). The server decodes Zoho bodies with UseNumber to
// preserve 64-bit ids in other Zoho modules (e.g. Mail folderId), so the CRM
// event parser must keep working under both decode modes.
func TestEventTimeStampNanoDecodeModes(t *testing.T) {
	t.Parallel()

	const wantNano = int64(1750102639787) * int64(1_000_000)

	firstEventTimestamp := func(t *testing.T, evt CollapsedSubscriptionEvent) int64 {
		t.Helper()

		events, err := evt.SubscriptionEventList()
		assert.NilError(t, err)
		assert.Equal(t, len(events), 1)

		nano, err := events[0].EventTimeStampNano()
		assert.NilError(t, err)

		return nano
	}

	t.Run("plain unmarshal (float64)", func(t *testing.T) {
		t.Parallel()

		var evt CollapsedSubscriptionEvent
		assert.NilError(t, json.Unmarshal([]byte(crmWebhookBody), &evt))

		assert.Equal(t, firstEventTimestamp(t, evt), wantNano)
	})

	t.Run("UseNumber (json.Number)", func(t *testing.T) {
		t.Parallel()

		var evt CollapsedSubscriptionEvent

		decoder := json.NewDecoder(bytes.NewReader([]byte(crmWebhookBody)))
		decoder.UseNumber()
		assert.NilError(t, decoder.Decode(&evt))

		assert.Equal(t, firstEventTimestamp(t, evt), wantNano)
	})
}
