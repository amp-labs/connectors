package attio

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// newTestEvent creates a SubscriptionEvent that mirrors a real Attio webhook
// payload: a top-level "events" array whose single element holds the event_type
// and an "id" object. The id values are stored as map[string]any because that is
// what encoding/json produces when decoding a real webhook body.
func newTestEvent(eventType string, idMap map[string]string) SubscriptionEvent {
	event := map[string]any{
		"event_type": eventType,
	}

	if idMap != nil {
		id := make(map[string]any, len(idMap))
		for k, v := range idMap {
			id[k] = v
		}

		event["id"] = id
	}

	return SubscriptionEvent{
		"events": []any{event},
	}
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
			name:        "Missing events key",
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
			name:        "Missing events key",
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
			name:        "Missing events key",
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
			name:        "Missing events key",
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
			name:        "Missing events key",
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

// TestSubscriptionEvent_RealWebhookPayload decodes a real Attio webhook body
// (via encoding/json, the way the server does) and verifies the parsing methods
// work end to end. This guards against the events-as-array / id map[string]any
// shape being mishandled.
func TestSubscriptionEvent_RealWebhookPayload(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"webhook_id": "wh-1",
		"events": [
			{
				"event_type": "note.updated",
				"id": {
					"workspace_id": "ws-1",
					"note_id": "note-9"
				}
			}
		]
	}`)

	var evt SubscriptionEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	eventType, err := evt.EventType()
	if err != nil {
		t.Fatalf("EventType() error: %v", err)
	}

	if eventType != common.SubscriptionEventTypeUpdate {
		t.Fatalf("EventType() = %v, want %v", eventType, common.SubscriptionEventTypeUpdate)
	}

	objectName, err := evt.ObjectName()
	if err != nil {
		t.Fatalf("ObjectName() error: %v", err)
	}

	if objectName != "note" {
		t.Fatalf("ObjectName() = %q, want %q", objectName, "note")
	}

	recordID, err := evt.RecordId()
	if err != nil {
		t.Fatalf("RecordId() error: %v", err)
	}

	if recordID != "note-9" {
		t.Fatalf("RecordId() = %q, want %q", recordID, "note-9")
	}

	workspace, err := evt.Workspace()
	if err != nil {
		t.Fatalf("Workspace() error: %v", err)
	}

	if workspace != "ws-1" {
		t.Fatalf("Workspace() = %q, want %q", workspace, "ws-1")
	}
}

func TestGetFieldNameFromObjectMetadata(t *testing.T) {
	t.Parallel()

	metadata := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"obj-uuid-1": {
				DisplayName: "Companies",
				Fields: map[string]common.FieldMetadata{
					"name": {
						DisplayName: "name",
						FieldId:     new("attr-uuid-1"),
					},
					"domains": {
						DisplayName: "domains",
						FieldId:     new("attr-uuid-2"),
					},
				},
			},
			"obj-uuid-2": {
				DisplayName: "People",
				Fields: map[string]common.FieldMetadata{
					"email": {
						DisplayName: "email",
						FieldId:     nil,
					},
				},
			},
		},
		Errors: map[string]error{},
	}

	tests := []struct {
		name        string
		objectID    string
		attributeID string
		expected    string
		expectedErr error
	}{
		{
			name:        "Found field by attribute ID",
			objectID:    "obj-uuid-1",
			attributeID: "attr-uuid-1",
			expected:    "name",
		},
		{
			name:        "Found second field by attribute ID",
			objectID:    "obj-uuid-1",
			attributeID: "attr-uuid-2",
			expected:    "domains",
		},
		{
			name:        "Object not found",
			objectID:    "unknown-obj",
			attributeID: "attr-uuid-1",
			expectedErr: common.ErrNotFound,
		},
		{
			name:        "Attribute not found in object",
			objectID:    "obj-uuid-1",
			attributeID: "unknown-attr",
			expectedErr: common.ErrNotFound,
		},
		{
			name:        "Nil FieldId is skipped",
			objectID:    "obj-uuid-2",
			attributeID: "any-attr",
			expectedErr: common.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := GetFieldNameFromObjectMetadata(metadata, tt.objectID, tt.attributeID)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}

				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
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

func TestGetObjectNameFromObjectMetadata(t *testing.T) {
	t.Parallel()

	metadata := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"obj-uuid-1": {
				DisplayName: "Companies",
			},
			"obj-uuid-2": {
				DisplayName: "People",
			},
		},
		Errors: map[string]error{},
	}

	tests := []struct {
		name        string
		objectID    string
		expected    string
		expectedErr error
	}{
		{
			name:     "Found object display name",
			objectID: "obj-uuid-1",
			expected: "Companies",
		},
		{
			name:     "Found second object display name",
			objectID: "obj-uuid-2",
			expected: "People",
		},
		{
			name:        "Object not found",
			objectID:    "unknown-obj",
			expectedErr: common.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := GetObjectNameFromObjectMetadata(metadata, tt.objectID)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}

				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
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
