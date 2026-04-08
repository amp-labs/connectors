package webhook

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestEvent(t *testing.T) {
	for _, tt := range []testroutines.SubscriptionEventTestCase{
		{
			Name: "Created event",
			Input: Event{
				"changeType": "created",
				"resource":   "Users/123",
				"resourceData": map[string]any{
					"id": "123",
				},
			},
			Expected: testroutines.SubscriptionEventExpected{
				EventType:    common.SubscriptionEventTypeCreate,
				RawEventName: "created",
				ObjectName:   "Users",
				RecordId:     "123",
			},
		},
		{
			Name: "Updated event",
			Input: Event{
				"changeType": "updated",
				"resource":   "Messages/456",
				"resourceData": map[string]any{
					"@odata.type": "#Microsoft.Graph.Message",
					"id":          "456",
				},
			},
			Expected: testroutines.SubscriptionEventExpected{
				EventType:    common.SubscriptionEventTypeUpdate,
				RawEventName: "updated",
				ObjectName:   "Messages",
				RecordId:     "456",
			},
		},
		{
			Name: "Deleted event",
			Input: Event{
				"changeType": "deleted",
				"resource":   "Users/789",
			},
			Expected: testroutines.SubscriptionEventExpected{
				EventType:    common.SubscriptionEventTypeDelete,
				RawEventName: "deleted",
				ObjectName:   "Users",
				RecordId:     "789",
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}
