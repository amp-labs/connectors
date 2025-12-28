package stripe

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

//nolint:funlen
func TestCollapsedSubscriptionEvent_Created(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscribe/setup_intent-created.json")

	var evt CollapsedSubscriptionEvent

	err := json.Unmarshal(data, &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	rawMap, err := evt.RawMap()
	assert.NilError(t, err, "RawMap should not return error")
	assert.Assert(t, rawMap != nil, "RawMap should not be nil")

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err, "SubscriptionEventList should not return error")
	assert.Equal(t, len(events), 1, "should have exactly one event")

	subEvt := events[0]

	eventType, err := subEvt.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, common.SubscriptionEventTypeCreate, "EventType should be Create")

	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, "setup_intent.created", "RawEventName should be setup_intent.created")

	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, "setup_intent", "ObjectName should be setup_intent")

	recordID, err := subEvt.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, "seti_1NG8Du2eZvKYlo2C9XMqbR0x", "RecordId should be seti_1NG8Du2eZvKYlo2C9XMqbR0x")

	workspace, err := subEvt.Workspace()
	assert.NilError(t, err, "Workspace should not return error")
	assert.Equal(t, workspace, "", "Workspace should be empty")

	timestamp, err := subEvt.EventTimeStampNano()
	assert.NilError(t, err, "EventTimeStampNano should not return error")
	assert.Assert(t, timestamp > 0, "EventTimeStampNano should be positive")

	updateEvt, ok := subEvt.(common.SubscriptionUpdateEvent)
	assert.Assert(t, ok, "should implement SubscriptionUpdateEvent")

	fields, err := updateEvt.UpdatedFields()
	assert.NilError(t, err, "UpdatedFields should not return error")
	assert.Equal(t, len(fields), 0, "created events should have no updated fields")
}
