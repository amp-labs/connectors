package attio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// newTestEvent creates a single SubscriptionEvent as produced by
// CollapsedSubscriptionEvent.SubscriptionEventList: one event object holding the
// event_type and an "id" object. The id values are stored as map[string]any
// because that is what encoding/json produces when decoding a real webhook body.
func newTestEvent(eventType string, idMap map[string]string) SubscriptionEvent {
	event := SubscriptionEvent{
		"event_type": eventType,
	}

	if idMap != nil {
		id := make(map[string]any, len(idMap))
		for k, v := range idMap {
			id[k] = v
		}

		event["id"] = id
	}

	return event
}

func TestSubscriptionEvent_EventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		event       SubscriptionEvent
		expected    common.SubscriptionEventType
		expectedErr bool
	}{
		{
			name:     "Created event",
			event:    newTestEvent("record.created", nil),
			expected: common.SubscriptionEventTypeCreate,
		},
		{
			name:     "Updated event",
			event:    newTestEvent("note.updated", nil),
			expected: common.SubscriptionEventTypeUpdate,
		},
		{
			name:     "Deleted event",
			event:    newTestEvent("record.deleted", nil),
			expected: common.SubscriptionEventTypeDelete,
		},
		{
			name:     "Unknown action maps to Other",
			event:    newTestEvent("record.merged", nil),
			expected: common.SubscriptionEventTypeOther,
		},
		{
			name:        "Empty event",
			event:       SubscriptionEvent{},
			expectedErr: true,
		},
		{
			name:        "Single part event type",
			event:       newTestEvent("invalid", nil),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.event.EventType()
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSubscriptionEvent_ObjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		event       SubscriptionEvent
		expected    string
		expectedErr bool
	}{
		{
			name:     "Extracts object name from event type",
			event:    newTestEvent("note.updated", nil),
			expected: "note",
		},
		{
			name:     "Extracts record object name",
			event:    newTestEvent("record.created", nil),
			expected: "record",
		},
		{
			name:        "Empty event",
			event:       SubscriptionEvent{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.event.ObjectName()
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptionEvent_RawEventName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		event       SubscriptionEvent
		expected    string
		expectedErr bool
	}{
		{
			name:     "Returns full event type string",
			event:    newTestEvent("note.updated", nil),
			expected: "note.updated",
		},
		{
			name:     "Returns created event type",
			event:    newTestEvent("record.created", nil),
			expected: "record.created",
		},
		{
			name:        "Empty event",
			event:       SubscriptionEvent{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.event.RawEventName()
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptionEvent_RecordId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		event       SubscriptionEvent
		expected    string
		expectedErr bool
	}{
		{
			name: "note event uses note_id",
			event: newTestEvent("note.updated", map[string]string{
				"workspace_id": "ws-123",
				"note_id":      "note-456",
			}),
			expected: "note-456",
		},
		{
			// note-content events carry note_id, not "note-content_id".
			// Ref: https://docs.attio.com/rest-api/webhook-reference/note-content-events/note-contentupdated
			name: "note-content event uses note_id",
			event: newTestEvent("note-content.updated", map[string]string{
				"workspace_id": "ws-123",
				"note_id":      "note-789",
			}),
			expected: "note-789",
		},
		{
			// workspace-member events carry workspace_member_id (underscore).
			// Ref: https://docs.attio.com/rest-api/webhook-reference/workspace-member-events/workspace-membercreated
			name: "workspace-member event uses workspace_member_id",
			event: newTestEvent("workspace-member.created", map[string]string{
				"workspace_id":        "ws-123",
				"workspace_member_id": "wm-001",
			}),
			expected: "wm-001",
		},
		{
			// record events carry record_id (alongside object_id).
			// Ref: https://docs.attio.com/rest-api/webhook-reference/record-events/recordcreated
			name: "record event uses record_id",
			event: newTestEvent("record.created", map[string]string{
				"workspace_id": "ws-123",
				"object_id":    "obj-1",
				"record_id":    "rec-001",
			}),
			expected: "rec-001",
		},
		{
			name: "Missing ID key for object",
			event: newTestEvent("note.updated", map[string]string{
				"workspace_id": "ws-123",
			}),
			expectedErr: true,
		},
		{
			name: "Unmapped event object returns error",
			event: newTestEvent("unknown-object.created", map[string]string{
				"workspace_id": "ws-123",
			}),
			expectedErr: true,
		},
		{
			name:        "Empty event",
			event:       SubscriptionEvent{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.event.RecordId()
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptionEvent_Workspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		event       SubscriptionEvent
		expected    string
		expectedErr bool
	}{
		{
			name: "Extracts workspace ID",
			event: newTestEvent("note.updated", map[string]string{
				"workspace_id": "ws-123",
				"note_id":      "note-456",
			}),
			expected: "ws-123",
		},
		{
			name: "Missing workspace_id",
			event: newTestEvent("note.updated", map[string]string{
				"note_id": "note-456",
			}),
			expectedErr: true,
		},
		{
			name:        "Empty event",
			event:       SubscriptionEvent{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.event.Workspace()
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptionEvent_RawMap(t *testing.T) {
	t.Parallel()

	evt := newTestEvent("note.updated", nil)

	result, err := evt.RawMap()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(result, map[string]any(evt)) {
		t.Fatal("RawMap should return a copy of the event map")
	}

	// Verify it's a clone (modifying result doesn't affect original).
	result["extra"] = "value"
	if _, exists := evt["extra"]; exists {
		t.Fatal("RawMap should return a clone, not the original")
	}
}

func TestSubscriptionEvent_UpdatedFields(t *testing.T) {
	t.Parallel()

	evt := newTestEvent("note.updated", nil)

	fields, err := evt.UpdatedFields()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fields) != 0 {
		t.Fatalf("expected empty fields, got %v", fields)
	}
}

func TestSubscriptionEvent_EventTimeStampNano(t *testing.T) {
	t.Parallel()

	evt := newTestEvent("note.updated", nil)

	_, err := evt.EventTimeStampNano()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestCollapsedSubscriptionEvent_RealWebhookPayload decodes a real Attio webhook
// body (via encoding/json, the way the server does), fans it out through
// SubscriptionEventList, and verifies the parsing methods work end to end on the
// resulting per-event SubscriptionEvents. It uses two events to also exercise the
// array fan-out (batching), guarding against the events-array / id map[string]any
// shape being mishandled.
func TestCollapsedSubscriptionEvent_RealWebhookPayload(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"webhook_id": "wh-1",
		"events": [
			{
				"event_type": "note.updated",
				"id": { "workspace_id": "ws-1", "note_id": "note-9" }
			},
			{
				"event_type": "record.created",
				"id": { "workspace_id": "ws-1", "object_id": "obj-1", "record_id": "rec-2" }
			}
		]
	}`)

	var collapsed CollapsedSubscriptionEvent
	if err := json.Unmarshal(body, &collapsed); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	events, err := collapsed.SubscriptionEventList()
	if err != nil {
		t.Fatalf("SubscriptionEventList() error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	cases := []struct {
		eventType  common.SubscriptionEventType
		objectName string
		recordID   string
	}{
		{common.SubscriptionEventTypeUpdate, "note", "note-9"},
		{common.SubscriptionEventTypeCreate, "record", "rec-2"},
	}

	for i, want := range cases {
		evt := events[i]

		eventType, err := evt.EventType()
		if err != nil {
			t.Fatalf("event %d EventType() error: %v", i, err)
		}

		if eventType != want.eventType {
			t.Fatalf("event %d EventType() = %v, want %v", i, eventType, want.eventType)
		}

		objectName, err := evt.ObjectName()
		if err != nil {
			t.Fatalf("event %d ObjectName() error: %v", i, err)
		}

		if objectName != want.objectName {
			t.Fatalf("event %d ObjectName() = %q, want %q", i, objectName, want.objectName)
		}

		recordID, err := evt.RecordId()
		if err != nil {
			t.Fatalf("event %d RecordId() error: %v", i, err)
		}

		if recordID != want.recordID {
			t.Fatalf("event %d RecordId() = %q, want %q", i, recordID, want.recordID)
		}

		workspace, err := evt.Workspace()
		if err != nil {
			t.Fatalf("event %d Workspace() error: %v", i, err)
		}

		if workspace != "ws-1" {
			t.Fatalf("event %d Workspace() = %q, want %q", i, workspace, "ws-1")
		}
	}
}

func TestCollapsedSubscriptionEvent_SubscriptionEventList_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event CollapsedSubscriptionEvent
	}{
		{
			name:  "missing events key",
			event: CollapsedSubscriptionEvent{"webhook_id": "wh-1"},
		},
		{
			name:  "events is not an array",
			event: CollapsedSubscriptionEvent{"events": map[string]any{"event_type": "note.updated"}},
		},
		{
			name:  "event element is not an object",
			event: CollapsedSubscriptionEvent{"events": []any{"not-an-object"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := tt.event.SubscriptionEventList(); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestConnector_GetObjectNameFromTypeId(t *testing.T) {
	t.Parallel()

	objectsResponse := testutils.DataFromFile(t, "objects.json")

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/v2/objects"),
		Then:  mockserver.Response(http.StatusOK, objectsResponse),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	tests := []struct {
		name        string
		event       SubscriptionEvent
		expected    string
		expectedErr bool
	}{
		{
			// record.* event: object_id resolved via GET /v2/objects.
			name: "record event resolves object_id to api_slug",
			event: newTestEvent("record.created", map[string]string{
				"workspace_id": "ws-1",
				"object_id":    "0e80364d-70b1-44d3-b7ba-0a6a564a7152",
				"record_id":    "rec-1",
			}),
			expected: "people",
		},
		{
			// core-object event: no object_id, falls back to ObjectName() (no fetch).
			name: "core event falls back to event_type object",
			event: newTestEvent("note.updated", map[string]string{
				"workspace_id": "ws-1",
				"note_id":      "note-1",
			}),
			expected: "note",
		},
		{
			name: "unknown object_id returns error",
			event: newTestEvent("record.created", map[string]string{
				"workspace_id": "ws-1",
				"object_id":    "does-not-exist",
				"record_id":    "rec-1",
			}),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := conn.GetObjectNameFromTypeId(t.Context(), tt.event)
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConnector_GetObjectNameFromTypeId_CachesOnConnector(t *testing.T) {
	t.Parallel()

	objectsResponse := testutils.DataFromFile(t, "objects.json")

	var calls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/objects" {
			atomic.AddInt32(&calls, 1)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(objectsResponse)
	}))
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// The same connector caches the object list, so repeated resolutions fetch once.
	for range 3 {
		name, err := conn.GetObjectNameFromTypeId(t.Context(), newTestEvent("record.created", map[string]string{
			"object_id": "bb3380d7-06a7-4948-9d62-3e735e782c5c",
			"record_id": "rec-1",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if name != "companies" {
			t.Fatalf("expected %q, got %q", "companies", name)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected objects to be fetched once with cache, got %d", got)
	}
}
