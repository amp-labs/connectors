package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSubscriptionEvent(t *testing.T) {
	t.Parallel()

	for _, tt := range []testconn.TestCaseSubscriptionEvent{
		{
			Name: "Unsupported event",
			Input: SubscriptionEvent{
				"subscriptionType": "someObject.creation",
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					RawEventName: "someObject.creation",
					EventType:    "create",
				},
				Err: testconn.SubscriptionEventExpectedErr{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType: "other",
				},
				Err: testconn.SubscriptionEventExpectedErr{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:    "create",
					RawEventName: "importantContacts.creation",
					ObjectName:   "contact",
				},
				Err: testconn.SubscriptionEventExpectedErr{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
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
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          "delete",
					RawEventName:       "contact.deletion",
					ObjectName:         "contact",
					Workspace:          "101",
					RecordId:           "789",
					EventTimeStampNano: 1625097600000000000,
				},
			}},
		},
		{
			Name: "Contact merge event resolves the surviving record id",
			Input: SubscriptionEvent{
				"subscriptionType": "contact.merge",
				"objectId":         123,
				"primaryObjectId":  123,
				"mergedObjectIds":  []any{float64(456), float64(789)},
				"newObjectId":      999,
				"portalId":         101,
				"occurredAt":       1625097600000,
			},
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          "merge",
					RawEventName:       "contact.merge",
					ObjectName:         "contact",
					Workspace:          "101",
					RecordId:           "999",
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

func TestSubscriptionEventMerge(t *testing.T) {
	t.Parallel()

	t.Run("PrimaryRecordId prefers newObjectId over primaryObjectId", func(t *testing.T) {
		t.Parallel()

		evt := SubscriptionEvent{
			"subscriptionType": "contact.merge",
			"objectId":         123,
			"primaryObjectId":  123,
			"newObjectId":      999,
		}

		id, err := evt.PrimaryRecordId()
		if err != nil {
			t.Fatalf("PrimaryRecordId returned error: %v", err)
		}

		if id != "999" {
			t.Errorf("PrimaryRecordId = %q, want %q", id, "999")
		}
	})

	t.Run("PrimaryRecordId falls back to primaryObjectId when no new record is created", func(t *testing.T) {
		t.Parallel()

		evt := SubscriptionEvent{
			"subscriptionType": "contact.merge",
			"objectId":         123,
			"primaryObjectId":  123,
		}

		id, err := evt.PrimaryRecordId()
		if err != nil {
			t.Fatalf("PrimaryRecordId returned error: %v", err)
		}

		if id != "123" {
			t.Errorf("PrimaryRecordId = %q, want %q", id, "123")
		}
	})

	t.Run("PrimaryRecordId falls back to objectId as a last resort", func(t *testing.T) {
		t.Parallel()

		evt := SubscriptionEvent{
			"subscriptionType": "contact.merge",
			"objectId":         123,
		}

		id, err := evt.PrimaryRecordId()
		if err != nil {
			t.Fatalf("PrimaryRecordId returned error: %v", err)
		}

		if id != "123" {
			t.Errorf("PrimaryRecordId = %q, want %q", id, "123")
		}
	})

	t.Run("MergedRecordIds returns the consolidated record ids", func(t *testing.T) {
		t.Parallel()

		evt := SubscriptionEvent{
			"subscriptionType": "contact.merge",
			"mergedObjectIds":  []any{float64(456), float64(789)},
		}

		ids, err := evt.MergedRecordIds()
		if err != nil {
			t.Fatalf("MergedRecordIds returned error: %v", err)
		}

		want := []string{"456", "789"}
		if len(ids) != len(want) {
			t.Fatalf("MergedRecordIds = %v, want %v", ids, want)
		}

		for i := range want {
			if ids[i] != want[i] {
				t.Errorf("MergedRecordIds[%d] = %q, want %q", i, ids[i], want[i])
			}
		}
	})

	t.Run("MergedRecordIds errors when the field is missing", func(t *testing.T) {
		t.Parallel()

		evt := SubscriptionEvent{
			"subscriptionType": "contact.merge",
		}

		if _, err := evt.MergedRecordIds(); err == nil {
			t.Error("MergedRecordIds expected an error for a payload without mergedObjectIds")
		}
	})
}
