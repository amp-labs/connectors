package salesforce

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

func TestSubscriptionEventUpdate(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscription/update_user.json")

	event := SubscriptionEvent{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	eventType, err := event.EventType()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, eventType, common.SubscriptionEventTypeUpdate, "EventType should be Update")

	rawEventType, err := event.RawEventName()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, rawEventType, "UPDATE", "RawEventName should be UPDATE")
	objectName, err := event.ObjectName()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, objectName, "User", "ObjectName should be User")
	workspace, err := event.Workspace()
	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, workspace, "", "Workspace should be empty")

	fields, err := event.UpdatedFields()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, len(fields), 3, "should have one updated field")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "LastModifiedById", "second field name should be LastModifiedById")
	assert.Equal(t, fields[2], "FirstName", "third field name should be FirstName")
}

func TestSubscriptionEventProperties(t *testing.T) {
	t.Parallel()

	eventNewAccountData := testutils.DataFromFile(t, "subscription/new_account.json")

	changeEvent := CollapsedSubscriptionEvent{}
	if err := json.Unmarshal(eventNewAccountData, &changeEvent); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	events, err := changeEvent.SubscriptionEventList()
	assert.NilError(t, err, "error should be nil")

	if len(events) != 1 {
		t.Fatalf("failed to start a test, expected to have only one event")
	}

	event := events[0]

	eventType, err := event.EventType()
	validateSubEvent(t, err, eventType, common.SubscriptionEventTypeCreate, "EventType")

	rawEventType, err := event.RawEventName()
	validateSubEvent(t, err, rawEventType, "CREATE", "RawEventName")

	objectName, err := event.ObjectName()
	validateSubEvent(t, err, objectName, "Account", "ObjectName")

	workspace, err := event.Workspace()
	validateSubEvent(t, err, workspace, "", "Workspace")

	recordID, err := event.RecordId()
	validateSubEvent(t, err, recordID, "0015f00002J9YYEAA3", "RecordId")

	timestamp, err := event.EventTimeStampNano()
	validateSubEvent(t, err, timestamp, 1712693965000, "EventTimeStampNano")
}

func validateSubEvent[V any](t *testing.T, err error, actual, expected V, methodName string) {
	t.Helper()

	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, actual, expected, "method "+methodName)
}
