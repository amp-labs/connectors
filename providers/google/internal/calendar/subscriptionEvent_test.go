package calendar

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestSubscriptionEventEventType(t *testing.T) {
	t.Parallel()

	const updatedMin = "2026-06-17T00:00:00.000Z"

	tests := []struct {
		name    string
		event   SubscriptionEvent
		want    common.SubscriptionEventType
		wantErr bool
	}{
		{
			name:  "cancelled status is a delete",
			event: SubscriptionEvent{Status: "cancelled", Created: "2026-06-01T09:00:00.000Z", UpdatedMin: updatedMin},
			want:  common.SubscriptionEventTypeDelete,
		},
		{
			name:  "cancelled wins even when created within the window",
			event: SubscriptionEvent{Status: "cancelled", Created: "2026-06-18T09:00:00.000Z", UpdatedMin: updatedMin},
			want:  common.SubscriptionEventTypeDelete,
		},
		{
			name:  "created within the window is a create",
			event: SubscriptionEvent{Status: "confirmed", Created: "2026-06-18T09:00:00.000Z", UpdatedMin: updatedMin},
			want:  common.SubscriptionEventTypeCreate,
		},
		{
			name:  "created exactly at the window boundary is a create",
			event: SubscriptionEvent{Status: "confirmed", Created: updatedMin, UpdatedMin: updatedMin},
			want:  common.SubscriptionEventTypeCreate,
		},
		{
			name:  "created before the window is an update",
			event: SubscriptionEvent{Status: "confirmed", Created: "2026-06-01T09:00:00.000Z", UpdatedMin: updatedMin},
			want:  common.SubscriptionEventTypeUpdate,
		},
		{
			name:  "missing created falls back to update",
			event: SubscriptionEvent{Status: "confirmed", Created: "", UpdatedMin: updatedMin},
			want:  common.SubscriptionEventTypeUpdate,
		},
		{
			name:    "unparseable updatedMin errors",
			event:   SubscriptionEvent{Status: "confirmed", Created: "2026-06-18T09:00:00.000Z", UpdatedMin: "not-a-time"},
			want:    common.SubscriptionEventTypeOther,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.event.EventType()

			assert.Equal(t, got, tt.want)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestSubscriptionEventsFromRecords(t *testing.T) {
	t.Parallel()

	const updatedMin = "2026-06-17T00:00:00.000Z"

	rows := []common.ReadResultRow{
		{
			Id: "evt-confirmed-1",
			Raw: map[string]any{
				"id":      "evt-confirmed-1",
				"status":  "confirmed",
				"created": "2026-06-01T09:42:20.000Z", // created before the window: an update
				"updated": "2026-06-18T09:00:00.000Z",
			},
		},
		{
			Id: "evt-cancelled-1",
			Raw: map[string]any{
				"id":      "evt-cancelled-1",
				"status":  "cancelled", // cancelled: a delete
				"created": "2026-06-01T09:42:17.000Z",
				"updated": "2026-06-18T09:30:00.000Z",
			},
		},
		{
			// id/status/created only in the marshaled Fields (lowercased keys), not Raw.
			Fields: map[string]any{
				"id":      "evt-new-1",
				"status":  "confirmed",
				"created": "2026-06-18T11:00:00.000Z", // created inside the window: a create
			},
		},
	}

	events := SubscriptionEventsFromRecords(rows, updatedMin)
	assert.Equal(t, len(events), 3)

	// Every event should be tagged with the window.
	for _, evt := range events {
		assert.Equal(t, evt.UpdatedMin, updatedMin)
	}

	wantTypes := []common.SubscriptionEventType{
		common.SubscriptionEventTypeUpdate,
		common.SubscriptionEventTypeDelete,
		common.SubscriptionEventTypeCreate,
	}

	for i, evt := range events {
		got, err := evt.EventType()
		assert.NilError(t, err)
		assert.Equal(t, got, wantTypes[i], "event %d", i)
	}

	// Fields-only row still resolves its id and object name.
	id, err := events[2].RecordId()
	assert.NilError(t, err)
	assert.Equal(t, id, "evt-new-1")

	obj, err := events[0].ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, obj, objectNameEvents)
}
