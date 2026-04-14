package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

type SubscriptionEventExpected struct {
	EventType          common.SubscriptionEventType
	RawEventName       string
	ObjectName         string
	Workspace          string
	RecordId           string
	EventTimeStampNano int64
	UpdatedFields      []string
}

type SubscriptionEventTestCase struct {
	Name     string
	Input    common.SubscriptionEvent
	Expected SubscriptionEventExpected
}

func (c SubscriptionEventTestCase) Run(t *testing.T) {
	t.Helper()

	// Test EventType
	eventType, err := c.Input.EventType()
	assert.NilError(t, err, "EventType should not return error")
	assert.Equal(t, eventType, c.Expected.EventType, "EventType should match")

	// Test RawEventName
	rawEventName, err := c.Input.RawEventName()
	assert.NilError(t, err, "RawEventName should not return error")
	assert.Equal(t, rawEventName, c.Expected.RawEventName, "RawEventName should match")

	// Test ObjectName
	objectName, err := c.Input.ObjectName()
	assert.NilError(t, err, "ObjectName should not return error")
	assert.Equal(t, objectName, c.Expected.ObjectName, "ObjectName should match")

	// Test Workspace
	workspace, err := c.Input.Workspace()
	assert.NilError(t, err, "Workspace should not return error")
	assert.Equal(t, workspace, c.Expected.Workspace, "Workspace should match")

	// Test RecordId
	recordID, err := c.Input.RecordId()
	assert.NilError(t, err, "RecordId should not return error")
	assert.Equal(t, recordID, c.Expected.RecordId, "RecordId should match")

	// Test EventTimeStampNano
	timestamp, err := c.Input.EventTimeStampNano()
	assert.NilError(t, err, "EventTimeStampNano should not return error")
	if c.Expected.EventTimeStampNano != 0 {
		assert.Equal(t, timestamp, c.Expected.EventTimeStampNano, "EventTimeStampNano should match")
	} else {
		assert.Assert(t, timestamp >= 0, "EventTimeStampNano should be non-negative")
	}

	// Test UpdatedFields if expected
	if len(c.Expected.UpdatedFields) > 0 {
		updateEvt, ok := c.Input.(common.SubscriptionUpdateEvent)
		assert.Assert(t, ok, "should implement SubscriptionUpdateEvent")

		fields, err := updateEvt.UpdatedFields()
		assert.NilError(t, err, "UpdatedFields should not return error")
		assert.DeepEqual(t, fields, c.Expected.UpdatedFields)
	}
}
