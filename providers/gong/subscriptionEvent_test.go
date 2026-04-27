package gong

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

// TestSubscriptionEvent_CallCreated exercises the happy path against a real
// sanitized gong webhook fixture (see test/webhook-call-created.json).
// Covers RawMap, SubscriptionEventList, EventType, RawEventName, ObjectName,
// RecordId, Workspace, and EventTimeStampNano in one pass.
func TestSubscriptionEvent_CallCreated(t *testing.T) {
	t.Parallel()

	var evt CollapsedSubscriptionEvent

	raw := testutils.DataFromFile(t, "webhook-call-created.json")
	if err := json.Unmarshal(raw, &evt); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	rawMap, err := evt.RawMap()
	assert.NilError(t, err)
	assert.Assert(t, rawMap != nil)

	events, err := evt.SubscriptionEventList()
	assert.NilError(t, err)
	assert.Equal(t, len(events), 1)

	subEvt := events[0]

	eventType, err := subEvt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, eventType, common.SubscriptionEventTypeCreate)

	rawEventName, err := subEvt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, rawEventName, "callCreated")

	objectName, err := subEvt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, objectName, "Call")

	recordID, err := subEvt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, recordID, "4570250816631531513")

	workspace, err := subEvt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, workspace, "1007648505208900737")

	ts, err := subEvt.EventTimeStampNano()
	assert.NilError(t, err)
	assert.Equal(t, ts, int64(0))
}
