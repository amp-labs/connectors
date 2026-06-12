package subscriptionhelper

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSegmentSubscriptionEvents(t *testing.T) {
	tests := []struct {
		name       string
		prevEvents map[ObjectName]ObjectEvents
		newEvents  map[ObjectName]ObjectEvents
		expected   EventSegments
	}{
		{
			name:       "Empty events",
			prevEvents: nil,
			newEvents:  nil,
			expected:   EventSegments{},
		},
		{
			name:       "Create only",
			prevEvents: nil,
			newEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"email"},
				},
			},
			expected: EventSegments{
				ToCreate: map[ObjectName]ObjectEvents{
					"User": {
						Events:      []common.SubscriptionEventType{"create"},
						WatchFields: []string{"email"},
					},
				},
			},
		},
		{
			name: "Remove only",
			prevEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"email"},
				},
			},
			newEvents: nil,
			expected: EventSegments{
				ToRemove: map[ObjectName]ObjectEvents{
					"User": {
						Events:      []common.SubscriptionEventType{"create"},
						WatchFields: []string{"email"},
					},
				},
			},
		},
		{
			name: "Keep only",
			prevEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"email"},
				},
			},
			newEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"email"},
				},
			},
			expected: EventSegments{
				ToKeep: map[ObjectName]ObjectEvents{
					"User": {
						Events:      []common.SubscriptionEventType{"create"},
						WatchFields: []string{"email"},
					},
				},
			},
		},
		{
			name: "Update only - different Events",
			prevEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"email"},
				},
			},
			newEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"create", "update"},
					WatchFields: []string{"email"},
				},
			},
			expected: EventSegments{
				ToUpdate: map[ObjectName]ObjectEvents{
					"User": {
						Events:      []common.SubscriptionEventType{"create", "update"},
						WatchFields: []string{"email"},
					},
				},
			},
		},
		{
			name: "Update only - different WatchFields",
			prevEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"update"},
					WatchFields: []string{"email"},
				},
			},
			newEvents: map[ObjectName]ObjectEvents{
				"User": {
					Events:      []common.SubscriptionEventType{"update"},
					WatchFields: []string{"email", "name"},
				},
			},
			expected: EventSegments{
				ToUpdate: map[ObjectName]ObjectEvents{
					"User": {
						Events:      []common.SubscriptionEventType{"update"},
						WatchFields: []string{"email", "name"},
					},
				},
			},
		},
		{
			name: "Mixed operations",
			prevEvents: map[ObjectName]ObjectEvents{
				"ObjToKeep": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"field1"},
				},
				"ObjToUpdate": {
					Events:      []common.SubscriptionEventType{"update"},
					WatchFields: []string{"field2"},
				},
				"ObjToRemove": {
					Events:      []common.SubscriptionEventType{"delete"},
					WatchFields: []string{"field3"},
				},
			},
			newEvents: map[ObjectName]ObjectEvents{
				"ObjToKeep": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"field1"},
				},
				"ObjToUpdate": {
					Events:      []common.SubscriptionEventType{"update"},
					WatchFields: []string{"field2", "field2_updated"},
				},
				"ObjToCreate": {
					Events:      []common.SubscriptionEventType{"create"},
					WatchFields: []string{"field4"},
				},
			},
			expected: EventSegments{
				ToCreate: map[ObjectName]ObjectEvents{
					"ObjToCreate": {
						Events:      []common.SubscriptionEventType{"create"},
						WatchFields: []string{"field4"},
					},
				},
				ToKeep: map[ObjectName]ObjectEvents{
					"ObjToKeep": {
						Events:      []common.SubscriptionEventType{"create"},
						WatchFields: []string{"field1"},
					},
				},
				ToUpdate: map[ObjectName]ObjectEvents{
					"ObjToUpdate": {
						Events:      []common.SubscriptionEventType{"update"},
						WatchFields: []string{"field2", "field2_updated"},
					},
				},
				ToRemove: map[ObjectName]ObjectEvents{
					"ObjToRemove": {
						Events:      []common.SubscriptionEventType{"delete"},
						WatchFields: []string{"field3"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutils.NewCompareResult()
			got := SegmentSubscriptionEvents(tt.prevEvents, tt.newEvents)

			if tt.expected.ToCreate == nil {
				tt.expected.ToCreate = make(map[ObjectName]ObjectEvents)
			}
			if tt.expected.ToKeep == nil {
				tt.expected.ToKeep = make(map[ObjectName]ObjectEvents)
			}
			if tt.expected.ToUpdate == nil {
				tt.expected.ToUpdate = make(map[ObjectName]ObjectEvents)
			}
			if tt.expected.ToRemove == nil {
				tt.expected.ToRemove = make(map[ObjectName]ObjectEvents)
			}

			result.Assert("ToCreate", tt.expected.ToCreate, got.ToCreate)
			result.Assert("ToKeep", tt.expected.ToKeep, got.ToKeep)
			result.Assert("ToUpdate", tt.expected.ToUpdate, got.ToUpdate)
			result.Assert("ToRemove", tt.expected.ToRemove, got.ToRemove)

			result.Validate(t, tt.name)
		})
	}
}
