package attio

import (
	"errors"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

// newTestEvent creates a SubscriptionEvent for testing.
// The "events" value is a map (not an array) so that asMap() falls back
// to returning the raw SubscriptionEvent, which is what the methods expect
// when they call m.Get("events").
func newTestEvent(eventType string, idMap map[string]string) SubscriptionEvent {
	inner := map[string]any{
		"event_type": eventType,
	}

	if idMap != nil {
		inner["id"] = idMap
	}

	return SubscriptionEvent{
		"events": inner,
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
			name: "Extracts record ID using object name as key",
			event: newTestEvent("note.updated", map[string]string{
				"workspace_id": "ws-123",
				"note_id":      "note-456",
			}),
			expected: "note-456",
		},
		{
			name: "Missing ID key for object",
			event: newTestEvent("note.updated", map[string]string{
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

	_, err := evt.UpdatedFields()
	if err == nil {
		t.Fatal("expected error, got nil")
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

func TestGetFieldNameFromObjectMetadata(t *testing.T) {
	t.Parallel()

	metadata := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"obj-uuid-1": {
				DisplayName: "Companies",
				Fields: map[string]common.FieldMetadata{
					"name": {
						DisplayName: "name",
						FieldId:     goutils.Pointer("attr-uuid-1"),
					},
					"domains": {
						DisplayName: "domains",
						FieldId:     goutils.Pointer("attr-uuid-2"),
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
