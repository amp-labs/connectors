package webhook

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestEvent(t *testing.T) {
	responseMessageCreated := testutils.DataFromFileAs[EventCollection](t, "event-message-created.json")
	responseMessageUpdated := testutils.DataFromFileAs[EventCollection](t, "event-message-updated.json")
	responseMessageDeleted := testutils.DataFromFileAs[EventCollection](t, "event-message-deleted.json")

	for _, tt := range []testroutines.SubscriptionEventTestCase{
		{
			Name:  "Created event",
			Input: responseMessageCreated,
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
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
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
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
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
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
