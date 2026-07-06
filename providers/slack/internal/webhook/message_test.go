package webhook

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestEvent(t *testing.T) {
	responseConversationCreated := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "channel-created.json")
	responseConversationDeleted := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "channel-deleted.json")
	responseConversationArchived := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "channel-archived.json")

	for _, tt := range []testconn.TestCaseSubscriptionEvent{
		{
			Name:  "Created event",
			Input: responseConversationCreated,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
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
