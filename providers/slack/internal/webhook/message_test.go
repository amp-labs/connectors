package webhook

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestEvent(t *testing.T) {
	responseConversationCreated := testutils.DataFromFileAs[Event](t, "channel-created.json")
	responseConversationDeleted := testutils.DataFromFileAs[Event](t, "channel-deleted.json")
	responseConversationArchived := testutils.DataFromFileAs[Event](t, "channel-archived.json")

	for _, tt := range []testroutines.SubscriptionEventTestCase{
		{
			Name:  "Created event",
			Input: responseConversationCreated,
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:          "create",
					RawEventName:       "channel_created",
					ObjectName:         "conversations",
					Workspace:          "T0B9P1UVDBL",
					RecordId:           "C0B9N5F9ULE",
					EventTimeStampNano: 1781119867000000000,
				},
			}},
		},
		{
			Name:  "Deleted event",
			Input: responseConversationDeleted,
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:          "delete",
					RawEventName:       "channel_deleted",
					ObjectName:         "conversations",
					Workspace:          "T0B9P1UVDBL",
					RecordId:           "C0B9MM44P2A",
					EventTimeStampNano: 1781119850000000000,
				},
			}},
		},
		{
			Name:  "Archived event",
			Input: responseConversationArchived,
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:          "other",
					RawEventName:       "channel_archive",
					ObjectName:         "conversations",
					Workspace:          "T0B9P1UVDBL",
					RecordId:           "C0B9QHQ367L",
					EventTimeStampNano: 1781122615000000000,
				},
			}},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}
