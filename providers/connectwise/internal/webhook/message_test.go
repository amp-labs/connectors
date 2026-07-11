package webhook

import (
	"fmt"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestEvent(t *testing.T) {
	responseContactCreated := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "contact-create.json")
	responseContactUpdated := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "contact-update.json")
	responseContactDeleted := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "contact-delete.json")

	for _, tt := range []testconn.TestCaseSubscriptionEvent{
		{
			Name:  "Created event",
			Input: responseContactCreated,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "create",
					RawEventName: "added",
					ObjectName:   "contacts",
					RecordId:     "57960",
				},
			}},
		},
		{
			Name:  "Updated event",
			Input: responseContactUpdated,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "update",
					RawEventName: "updated",
					ObjectName:   "contacts",
					RecordId:     "57961",
				},
			}},
		},
		{
			Name:  "Deleted event",
			Input: responseContactDeleted,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "delete",
					RawEventName: "deleted",
					ObjectName:   "contacts",
					RecordId:     "57962",
				},
			}},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}

func TestEventRecord(t *testing.T) {
	t.Parallel()

	collapsed := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "contact-update.json")

	events, err := collapsed.SubscriptionEventList()
	if err != nil {
		t.Fatalf("SubscriptionEventList: %v", err)
	}

	withRecord, ok := events[0].(common.SubscriptionEventWithRecord)
	if !ok {
		t.Fatal("connectwise Event does not implement SubscriptionEventWithRecord")
	}

	record, err := withRecord.Record()
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	result := testutils.NewCompareResult()
	result.Assert("id", record["id"], float64(57961))
	result.Assert("firstName", record["firstName"], "Xzavier Sawayn")
	result.Validate(t, "inline Entity should unmarshal into the record")
}

func TestEventRecordMissingEntity(t *testing.T) {
	t.Parallel()

	if _, err := (Event{"Action": "updated", "ID": float64(1)}).Record(); err == nil {
		t.Fatal("expected error when 'Entity' is absent")
	}
}

func TestObjectTypeMappings(t *testing.T) {
	t.Parallel()

	result := testutils.NewCompareResult()

	// Both maps must be of the same length.
	result.Assert("map size", len(ObjectTypeToObjectName), len(ObjectNameToObjectType))

	for key, value := range ObjectNameToObjectType {
		result.Assert(fmt.Sprintf("pair [%v:%v]", key, value), key, ObjectTypeToObjectName[value])
	}

	result.Validate(t, "ObjectType mapping should be consistent both ways")
}
