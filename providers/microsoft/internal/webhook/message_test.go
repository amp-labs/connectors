package webhook

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestEvent(t *testing.T) {
	responseMessageCreated := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "event-message-created.json")
	responseMessageUpdated := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "event-message-updated.json")
	responseMessageDeleted := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "event-message-deleted.json")

	for _, tt := range []testconn.TestCaseSubscriptionEvent{
		{
			Name:  "Created event",
			Input: responseMessageCreated,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "create",
					RawEventName: "created",
					ObjectName:   "me/messages",
					RecordId:     "message_123",
				},
			}},
		},
		{
			Name:  "Updated event",
			Input: responseMessageUpdated,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "update",
					RawEventName: "updated",
					ObjectName:   "me/messages",
					RecordId:     "message_654",
				},
			}},
		},
		{
			Name:  "Deleted event",
			Input: responseMessageDeleted,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "delete",
					RawEventName: "deleted",
					ObjectName:   "me/messages",
					RecordId:     "message_798",
				},
			}},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}
