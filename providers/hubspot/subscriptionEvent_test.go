package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSubscriptionEvent(t *testing.T) {
	t.Parallel()

	for _, tt := range []testroutines.SubscriptionEventTestCase{
		{
			Name: "Unsupported event",
			Input: SubscriptionEvent{
				"subscriptionType": "someObject.creation",
			},
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					RawEventName: "someObject.creation",
					EventType:    "create",
				},
				Err: testroutines.SubscriptionEventExpectedErr{
					EventType:          nil,
					RawEventName:       nil,
					ObjectName:         testutils.StringError("subscription is not supported for the object 'someObject'"),
					Workspace:          testutils.StringError("key not found"),
					RecordId:           testutils.StringError("key not found"),
					EventTimeStampNano: testutils.StringError("key not found"),
				},
			}},
		},
		{
			Name: "Empty object name of the event",
			Input: SubscriptionEvent{
				"subscriptionType": "",
			},
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType: "other",
				},
				Err: testroutines.SubscriptionEventExpectedErr{
					EventType:          testutils.StringError("unexpected subscription event type: ''"),
					RawEventName:       nil,
					ObjectName:         testutils.StringError("subscription is not supported for the object ''"),
					Workspace:          testutils.StringError("key not found"),
					RecordId:           testutils.StringError("key not found"),
					EventTimeStampNano: testutils.StringError("key not found"),
				},
			}},
		},
		{
			Name: "Hubspot object type id is mapped to human readable object name",
			Input: SubscriptionEvent{
				"objectTypeId":     "0-1",
				"subscriptionType": "importantContacts.creation",
			},
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:    "create",
					RawEventName: "importantContacts.creation",
					ObjectName:   "contact",
				},
				Err: testroutines.SubscriptionEventExpectedErr{
					EventType:          nil,
					RawEventName:       nil,
					ObjectName:         nil,
					Workspace:          testutils.StringError("key not found"),
					RecordId:           testutils.StringError("key not found"),
					EventTimeStampNano: testutils.StringError("key not found"),
				},
			}},
		},
		{
			Name: "Contact creation event",
			Input: SubscriptionEvent{
				"subscriptionType": "contact.creation",
				"objectId":         123,
				"occurredAt":       1625097600000,
				"portalId":         101,
			},
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:          "create",
					RawEventName:       "contact.creation",
					ObjectName:         "contact",
					RecordId:           "123",
					Workspace:          "101",
					EventTimeStampNano: 1625097600000000000,
				},
			}},
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
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:          "update",
					RawEventName:       "contact.propertyChange",
					ObjectName:         "contact",
					RecordId:           "456",
					Workspace:          "101",
					UpdatedFields:      []string{"email"},
					EventTimeStampNano: 1625097600000000000,
				},
			}},
		},
		{
			Name: "Contact deletion event",
			Input: SubscriptionEvent{
				"subscriptionType": "contact.deletion",
				"objectId":         789,
				"portalId":         101,
				"occurredAt":       1625097600000,
			},
			Expected: []testroutines.SubscriptionEventExpected{{
				Data: testroutines.SubscriptionEventExpectedData{
					EventType:          "delete",
					RawEventName:       "contact.deletion",
					ObjectName:         "contact",
					Workspace:          "101",
					RecordId:           "789",
					EventTimeStampNano: 1625097600000000000,
				},
			}},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}
