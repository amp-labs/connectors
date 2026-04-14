package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestSubscriptionEvent(t *testing.T) {
	t.Parallel()

	for _, tt := range []testroutines.SubscriptionEventTestCase{
		{
			Name: "Contact creation event",
			Input: SubscriptionEvent{
				"subscriptionType": "contact.creation",
				"objectId":         123,
				"occurredAt":       1625097600000,
				"portalId":         101,
			},
			Expected: testroutines.SubscriptionEventExpected{
				EventType:          common.SubscriptionEventTypeCreate,
				RawEventName:       "contact.creation",
				ObjectName:         "contact",
				RecordId:           "123",
				Workspace:          "101",
				EventTimeStampNano: 1625097600000000000,
			},
		},
		{
			Name: "Contact property change event",
			Input: SubscriptionEvent{
				"subscriptionType": "contact.propertyChange",
				"objectId":         456,
				"propertyName":     "email",
				"portalId":         101,
				"occurredAt":       1625097600000,
			},
			Expected: testroutines.SubscriptionEventExpected{
				EventType:          common.SubscriptionEventTypeUpdate,
				RawEventName:       "contact.propertyChange",
				ObjectName:         "contact",
				RecordId:           "456",
				Workspace:          "101",
				UpdatedFields:      []string{"email"},
				EventTimeStampNano: 1625097600000000000,
			},
		},
		{
			Name: "Contact deletion event",
			Input: SubscriptionEvent{
				"subscriptionType": "contact.deletion",
				"objectId":         789,
				"portalId":         101,
				"occurredAt":       1625097600000,
			},
			Expected: testroutines.SubscriptionEventExpected{
				EventType:          common.SubscriptionEventTypeDelete,
				RawEventName:       "contact.deletion",
				ObjectName:         "contact",
				RecordId:           "789",
				Workspace:          "101",
				EventTimeStampNano: 1625097600000000000,
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}
