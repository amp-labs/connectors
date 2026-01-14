package outreach

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

//nolint:funlen
func TestCollapsedSubscriptionEvent_Created(t *testing.T) {
	t.Parallel()

	evtStr := `{
		"data": {
			"type": "account",
			"id": 13,
			"attributes": {
				"createdAt": "2025-11-04T09:40:36.000Z",
				"updatedAt": "2025-11-04T09:40:36.000Z",
				"named": true,
				"domain": "test.com",
				"externalSource": "outreach-api",
				"name": "this is a test"
			},
			"relationships": {
				"owner": {
					"type": "owner",
					"id": 2
				}
			}
		},
		"meta": {
			"deliveredAt": "2025-11-04T01:40:36.795-08:00",
			"eventName": "account.created",
			"jobId": "13dc9ab5-5ccc-4fbb-bdf9-cdbcdd986621"
		}
	}`

	var evt CollapsedSubscriptionEvent

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	// Test RawMap
	rawMap, err := evt.RawMap()
	assert.NilError(t, err, "RawMap should not return error")
	assert.Assert(t, rawMap != nil, "RawMap should not be nil")

	// Test SubscriptionEventList
	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err, "SubscriptionEventList should not return error")
	assert.Equal(t, len(events), 1, "should have exactly one event")

	subEvt := events[0]

	// Test EventType
	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeCreate, "EventType should be Create")

	// Test RawEventName
	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "account.created", "RawEventName should be account.created")

	// Test ObjectName
	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "account", "ObjectName should be account")

	// Test RecordId
	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "13", "RecordId should be 13")

	// Test Workspace
	workspace, err := subEvt.Workspace()
	assert.NilError(t, err, "Workspace should not return error")
	assert.Equal(t, workspace, "", "Workspace should be empty")

	// Test EventTimeStampNano
	timestamp, err := subEvt.EventTimeStampNano()
	assert.NilError(t, err, "EventTimeStampNano should not return error")
	assert.Assert(t, timestamp > 0, "EventTimeStampNano should be positive")

	// Test UpdatedFields via type assertion
	updateEvt, ok := subEvt.(common.SubscriptionUpdateEvent)
	assert.Assert(t, ok, "should implement SubscriptionUpdateEvent")

	fields, err := updateEvt.UpdatedFields()
	assert.NilError(t, err, "UpdatedFields should not return error")
	assert.Assert(t, len(fields) > 0, "should have updated fields")
}

//nolint:funlen
func TestCollapsedSubscriptionEvent_Updated(t *testing.T) {
	t.Parallel()

	evtStr := `{
		"data": {
			"type": "prospect",
			"id": 456,
			"attributes": {
				"updatedAt": "2025-11-05T10:30:00.000Z",
				"firstName": "John",
				"lastName": "Doe"
			}
		},
		"meta": {
			"deliveredAt": "2025-11-05T02:30:00.123-08:00",
			"eventName": "prospect.updated",
			"jobId": "abc123-def456"
		}
	}`

	var evt CollapsedSubscriptionEvent

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err, "SubscriptionEventList should not return error")
	assert.Equal(t, len(events), 1, "should have exactly one event")

	subEvt := events[0]

	// Test EventType
	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeUpdate, "EventType should be Update")

	// Test RawEventName
	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "prospect.updated", "RawEventName should be prospect.updated")

	// Test ObjectName
	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "prospect", "ObjectName should be prospect")

	// Test RecordId
	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "456", "RecordId should be 456")

	// Test UpdatedFields
	updateEvt, ok := subEvt.(common.SubscriptionUpdateEvent)
	assert.Assert(t, ok, "should implement SubscriptionUpdateEvent")

	fields, err := updateEvt.UpdatedFields()
	assert.NilError(t, err, "UpdatedFields should not return error")
	assert.Equal(t, len(fields), 3, "should have 3 updated fields")
}

func TestCollapsedSubscriptionEvent_Destroyed(t *testing.T) {
	t.Parallel()

	// Per Outreach docs: "deleted actions will contain the last known set of attributes"
	// Action is "destroyed" not "deleted"
	evtStr := `{
		"data": {
			"type": "task",
			"id": 789,
			"attributes": {
				"completed": true,
				"completedAt": "2025-11-06T14:00:00.000Z",
				"state": "completed"
			}
		},
		"meta": {
			"deliveredAt": "2025-11-06T07:00:00.000-08:00",
			"eventName": "task.destroyed",
			"jobId": "destroy-job-123"
		}
	}`

	var evt CollapsedSubscriptionEvent

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err, "SubscriptionEventList should not return error")
	assert.Equal(t, len(events), 1, "should have exactly one event")

	subEvt := events[0]

	// Test EventType
	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeDelete, "EventType should be Delete")

	// Test RawEventName
	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "task.destroyed", "RawEventName should be task.destroyed")

	// Test ObjectName
	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "task", "ObjectName should be task")

	// Test RecordId
	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "789", "RecordId should be 789")
}

func TestSubscriptionEvent_NumericId(t *testing.T) {
	t.Parallel()

	// Test that numeric IDs are correctly converted to strings
	evtStr := `{
		"data": {
			"type": "sequence",
			"id": 12345,
			"attributes": {}
		},
		"meta": {
			"deliveredAt": "2025-11-07T12:00:00.000Z",
			"eventName": "sequence.created",
			"jobId": "test-job"
		}
	}`

	var evt SubscriptionEvent

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	recordID, err := evt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "12345", "RecordId should be string '12345'")
}
