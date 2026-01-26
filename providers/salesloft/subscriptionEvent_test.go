package salesloft

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

//nolint:funlen,cyclop
func TestCollapsedSubscriptionEvent_Created(t *testing.T) {
	t.Parallel()

	evtStr := `{
  "_integration_id": null,
  "_integration_name": null,
  "_integration_step_id": null,
  "_integration_task_id": null,
  "_integration_task_type_label": null,
  "completed_at": null,
  "completed_by": null,
  "created_at": "2026-01-22T01:22:45.587713-05:00",
  "created_by_user": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  },
  "current_state": "scheduled",
  "custom_attribute_resources": {},
  "custom_attributes": {},
  "description": "Call John to discuss about the projects",
  "due_at": null,
  "due_date": "2026-01-31",
  "expires_after": null,
  "id": 721356732,
  "instigator": {
    "action_caller_id": 49067,
    "action_caller_name": "Int User",
    "metadata": {},
    "reason": "api",
    "type": "manual",
    "user_guid": "0863ed13-7120-479b-8650-206a3679e2fb"
  },
  "multitouch_group_id": null,
  "object_references": [],
  "person": {
    "_href": "https://api.salesloft.com/v2/people/436664215",
    "id": 436664215
  },
  "remind_at": null,
  "reminded": false,
  "rollback_reason": null,
  "score": {
    "factors": {},
    "prioritizer_uuid": "salesloft.prioritizers/rhythm",
    "score": "0.0"
  },
  "source": "salesloft.api",
  "subject": "Follow-up with John Kelly",
  "task_type": "general",
  "updated_at": "2026-01-22T01:22:45.587713-05:00",
  "user": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  }
}
`

	var evt CollapsedSubscriptionEvent

	// Create mock request with proper header initialization
	mockRequest := &http.Request{
		Header: make(http.Header),
	}
	mockRequest.Header.Set("x-salesloft-event", "task_created")

	preLoadedData := &common.SubscriptionEventPreLoadData{
		Request: mockRequest,
	}

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

	// Call PreLoadData to populate the eventHeader field
	err = subEvt.PreLoadData(preLoadedData)
	assert.NilError(t, err, "PreLoadData should not return error")

	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeCreate, "EventType should be Create")

	// Test RawEventName
	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "task_created", "RawEventName should be task_created")

	// Test ObjectName
	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "tasks", "ObjectName should be tasks")

	// Test RecordId
	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "721356732", "RecordId should be 721356732")

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
	assert.Error(t, err, "updated fields are not supported by Salesloft webhooks")
	assert.Assert(t, len(fields) == 0, "shouldnot have updated fields")
}

func TestCollapsedSubscriptionEvent_Updated(t *testing.T) {
	t.Parallel()

	evtStr := `{
  "_integration_id": null,
  "_integration_name": null,
  "_integration_step_id": null,
  "_integration_task_id": null,
  "_integration_task_type_label": null,
  "completed_at": null,
  "completed_by": null,
  "created_at": "2026-01-22T01:22:45.587713-05:00",
  "created_by_user": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  },
  "current_state": "scheduled",
  "custom_attribute_resources": {},
  "custom_attributes": {},
  "description": "Call John to discuss about the projects and this is updated to test update webhook event",
  "due_at": null,
  "due_date": "2026-01-31",
  "expires_after": null,
  "id": 721356732,
  "instigator": {
    "action_caller_id": 49067,
    "action_caller_name": "Int User",
    "metadata": {},
    "reason": "api",
    "type": "manual",
    "user_guid": "0863ed13-7120-479b-8650-206a3679e2fb"
  },
  "multitouch_group_id": null,
  "object_references": [],
  "person": {
    "_href": "https://api.salesloft.com/v2/people/436664215",
    "id": 436664215
  },
  "remind_at": null,
  "reminded": false,
  "rollback_reason": null,
  "score": {
    "factors": {},
    "prioritizer_uuid": "salesloft.prioritizers/rhythm",
    "score": "2.0"
  },
  "source": "salesloft.api",
  "subject": "Follow-up with John Kelly",
  "task_type": "general",
  "updated_at": "2026-01-22T02:31:27.782586-05:00",
  "user": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  }
}`

	var evt CollapsedSubscriptionEvent

	mockRequest := &http.Request{
		Header: make(http.Header),
	}

	mockRequest.Header.Set("x-salesloft-event", "task_updated")

	preLoadedData := common.SubscriptionEventPreLoadData{
		Request: mockRequest,
	}

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err, "SubscriptionEventList should not return error")
	assert.Equal(t, len(events), 1, "should have exactly one event")

	subEvt := events[0]

	err = subEvt.PreLoadData(&preLoadedData)
	assert.Assert(t, err, "PreLoadData should not return error")

	// Test EventType
	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeUpdate, "EventType should be Update")

	// Test RawEventName
	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "task_updated", "RawEventName should be task_updated")

	// Test ObjectName
	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "tasks", "ObjectName should be tasks")

	// Test RecordId
	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "721356732", "RecordId should be 721356732")

	// Test UpdatedFields via type assertion
	updateEvt, ok := subEvt.(common.SubscriptionUpdateEvent)
	assert.Assert(t, ok, "should implement SubscriptionUpdateEvent")

	fields, err := updateEvt.UpdatedFields()
	assert.Error(t, err, "updated fields are not supported by Salesloft webhooks")
	assert.Assert(t, len(fields) == 0, "shouldnot have updated fields")
}

func TestCollapsedSubscriptionEvent_Deleted(t *testing.T) {
	t.Parallel()

	evtStr := `
	{
  "custom_fields": {},
  "user_relationships": [],
  "account_tier": null,
  "last_contacted_at": null,
  "archived_at": null,
  "revenue_range": null,
  "description": null,
  "tags": [],
  "last_contacted_by": null,
  "id": 48371772,
  "counts": {
    "people": null
  },
  "created_at": "2024-06-07T10:50:43.502576-04:00",
  "twitter_handle": null,
  "name": "Puma",
  "linkedin_url": null,
  "company_type": null,
  "prospector_engagement_level": null,
  "owner_crm_id": null,
  "do_not_contact": false,
  "crm_url": null,
  "last_contacted_type": null,
  "street": null,
  "size": null,
  "creator": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  },
  "industry": null,
  "locale": null,
  "prospector_engagement_score": null,
  "crm_object_type": "account",
  "founded": null,
  "website": null,
  "crm_id": null,
  "company_stage": null,
  "country": null,
  "owner": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  },
  "city": null,
  "conversational_name": null,
  "state": null,
  "domain": "https://us.puma.com/us/en",
  "postal_code": null,
  "phone": null,
  "last_contacted_person": null,
  "updated_at": "2026-01-22T02:45:21.340083-05:00"
}	
	`

	var evt CollapsedSubscriptionEvent

	mockRequest := &http.Request{
		Header: make(http.Header),
	}

	mockRequest.Header.Set("x-salesloft-event", "account_deleted")

	preLoadedData := common.SubscriptionEventPreLoadData{
		Request: mockRequest,
	}

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err, "SubscriptionEventList should not return error")
	assert.Equal(t, len(events), 1, "should have exactly one event")

	subEvt := events[0]

	err = subEvt.PreLoadData(&preLoadedData)
	assert.Assert(t, err, "PreLoadData should not return error")

	// Test EventType
	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeDelete, "EventType should be Delete")

	// Test RawEventName
	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "account_deleted", "RawEventName should be account_deleted")

	// Test ObjectName
	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "accounts", "ObjectName should be accounts")

	// Test RecordId
	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "48371772", "RecordId should be 48371772")

	// Test UpdatedFields via type assertion
	updateEvt, ok := subEvt.(common.SubscriptionUpdateEvent)
	assert.Assert(t, ok, "should implement SubscriptionUpdateEvent")

	fields, err := updateEvt.UpdatedFields()
	assert.Error(t, err, "updated fields are not supported by Salesloft webhooks")
	assert.Assert(t, len(fields) == 0, "shouldnot have updated fields")
}
