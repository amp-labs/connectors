package webhook

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestEvent(t *testing.T) {
	responseConversationCreated := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "channel-created.json")
	responseConversationDeleted := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "channel-deleted.json")
	responseConversationArchived := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "channel-archived.json")
	responseMessageReply := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "message-reply.json")
	responseMessageChanged := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "message-changed.json")
	responseMessageDeleted := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, "message-deleted.json")

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
		{
			// A thread reply is a plain "message" event carrying a "thread_ts";
			// the record id is the composite "<channel>:<ts>" pair.
			Name:  "Message reply event",
			Input: responseMessageReply,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          "create",
					RawEventName:       "message",
					ObjectName:         "messages",
					Workspace:          "T0B9P1UVDBL",
					RecordId:           "C0B9N5F9ULE:1781123456.000200",
					EventTimeStampNano: 1781123456000000000,
				},
			}},
		},
		{
			// The event name comes from the "subtype"; the record id points at the
			// edited message ("event.message.ts"), not the time of the edit.
			Name:  "Message changed event",
			Input: responseMessageChanged,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          "update",
					RawEventName:       "message_changed",
					ObjectName:         "messages",
					Workspace:          "T0B9P1UVDBL",
					RecordId:           "C0B9N5F9ULE:1781123456.000200",
					EventTimeStampNano: 1781123500000000000,
				},
			}},
		},
		{
			// The record id points at the deleted message ("event.deleted_ts"),
			// not the time of the deletion.
			Name:  "Message deleted event",
			Input: responseMessageDeleted,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          "delete",
					RawEventName:       "message_deleted",
					ObjectName:         "messages",
					Workspace:          "T0B9P1UVDBL",
					RecordId:           "C0B9N5F9ULE:1781123456.000200",
					EventTimeStampNano: 1781123600000000000,
				},
			}},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)
		})
	}
}

// subscriptionEventFromFile parses the fixture and returns the single event produced by
// SubscriptionEventList.
func subscriptionEventFromFile(t *testing.T, fileName string) common.SubscriptionEvent {
	t.Helper()

	collapsed := testutils.DataFromFileAs[CollapsedSubscriptionEvent](t, fileName)

	events, err := collapsed.SubscriptionEventList()
	if err != nil {
		t.Fatalf("SubscriptionEventList: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %v", len(events))
	}

	return events[0]
}

func TestMessageEventRecord(t *testing.T) {
	t.Parallel()

	event := subscriptionEventFromFile(t, "message-reply.json")

	withRecord, ok := event.(common.SubscriptionEventWithRecord)
	if !ok {
		t.Fatal("slack message event does not implement SubscriptionEventWithRecord")
	}

	row, err := withRecord.Record([]string{"text", "thread_ts"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	result := testutils.NewCompareResult()
	// Id is the composite "<channel>:<ts>" identifier, matching RecordId().
	result.Assert("id", row.Id, "C0B9N5F9ULE:1781123456.000200")
	// Fields is the requested subset, with lowercased keys.
	result.Assert("fields.text", row.Fields["text"], "Looping back on this thread")
	result.Assert("fields.thread_ts", row.Fields["thread_ts"], "1781100000.000100")
	// Fields that were not requested are excluded from Fields.
	result.Assert("fields.user absent", row.Fields["user"], nil)
	// Raw carries the full message from the payload.
	result.Assert("raw.channel", row.Raw["channel"], "C0B9N5F9ULE")
	result.Assert("raw.user", row.Raw["user"], "U0B9R8LFG1F")
	result.Validate(t, "inline message should marshal into a ReadResultRow like a read")
}

func TestMessageEventRecordChanged(t *testing.T) {
	t.Parallel()

	event := subscriptionEventFromFile(t, "message-changed.json")

	withRecord, ok := event.(common.SubscriptionEventWithRecord)
	if !ok {
		t.Fatal("slack message event does not implement SubscriptionEventWithRecord")
	}

	row, err := withRecord.Record([]string{"text"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	result := testutils.NewCompareResult()
	result.Assert("id", row.Id, "C0B9N5F9ULE:1781123456.000200")
	// The record is the updated message from "event.message".
	result.Assert("fields.text", row.Fields["text"], "Looping back on this thread (edited)")
	// The channel lives on the envelope only and is copied onto the record.
	result.Assert("raw.channel", row.Raw["channel"], "C0B9N5F9ULE")
	result.Validate(t, "message_changed should surface the updated message as the record")
}

func TestMessageEventRecordDeleted(t *testing.T) {
	t.Parallel()

	event := subscriptionEventFromFile(t, "message-deleted.json")

	withRecord, ok := event.(common.SubscriptionEventWithRecord)
	if !ok {
		t.Fatal("slack message event does not implement SubscriptionEventWithRecord")
	}

	row, err := withRecord.Record([]string{"text"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	if len(row.Raw) != 0 {
		t.Fatalf("expected no record body for a deleted message, got %v", row.Raw)
	}
}

// Non-message events must stay plain Events: they do not carry a full record inline, so
// they must not advertise SubscriptionEventWithRecord (the record is fetched by id).
func TestNonMessageEventCarriesNoInlineRecord(t *testing.T) {
	t.Parallel()

	event := subscriptionEventFromFile(t, "channel-created.json")

	if _, ok := event.(common.SubscriptionEventWithRecord); ok {
		t.Fatal("non-message slack events must not implement SubscriptionEventWithRecord")
	}
}
