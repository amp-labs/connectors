package salesforce

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// The fixtures below are real outbound messages captured from a Salesforce dev org
// with test/salesforce/subscribe/flow-create. Only the org id, instance host,
// action ids and record ids are swapped for the placeholders Salesforce's own docs
// use — element order, namespaces, indentation and timestamp formats are exactly as
// Salesforce delivered them, which is the part worth pinning.
const (
	fixtureCreateAccount       = "subscription/outbound_message_create_account.xml"
	fixtureBatchUpdateContacts = "subscription/outbound_message_batch_update_contacts.xml"

	// A matched pair: the same Account, under the same outbound message (so the
	// same requested field list), captured before and after Industry was set.
	fixtureNullFieldOmitted = "subscription/outbound_message_null_field_omitted.xml"
	fixtureFieldPopulated   = "subscription/outbound_message_field_populated.xml"
)

func parseFixtureEvents(t *testing.T, fixture string) []common.SubscriptionEvent {
	t.Helper()

	envelope, err := ParseOutboundMessage(testutils.DataFromFile(t, fixture))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	events, err := envelope.SubscriptionEventList()
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}

	return events
}

// TestParseOutboundMessageCreate covers a real single-record notification. The
// create/update distinction is inferred from the audit timestamps, since outbound
// messages carry no event type: equal CreatedDate and LastModifiedDate means the
// save that fired the flow created the record.
func TestParseOutboundMessageCreate(t *testing.T) {
	t.Parallel()

	events := parseFixtureEvents(t, fixtureCreateAccount)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	objectName, err := event.ObjectName()
	if err != nil || objectName != "Account" {
		t.Errorf("ObjectName() = %q, %v; want Account", objectName, err)
	}

	recordID, err := event.RecordId()
	if err != nil || recordID != "001xx000003DGb2AAG" {
		t.Errorf("RecordId() = %q, %v; want 001xx000003DGb2AAG", recordID, err)
	}

	eventType, err := event.EventType()
	if err != nil || eventType != common.SubscriptionEventTypeCreate {
		t.Errorf("EventType() = %q, %v; want create", eventType, err)
	}

	rawName, err := event.RawEventName()
	if err != nil || rawName != "CREATE" {
		t.Errorf("RawEventName() = %q, %v; want CREATE", rawName, err)
	}

	// 2026-09-08T23:40:33.000Z
	nano, err := event.EventTimeStampNano()
	if err != nil || nano != int64(1788910833000000000) {
		t.Errorf("EventTimeStampNano() = %d, %v; want 1788910833000000000", nano, err)
	}
}

// TestParseOutboundMessageExpandsBatchedNotifications covers the batching Salesforce
// actually performs: this fixture is one POST produced by a single bulk edit of three
// Contacts. Salesforce packs up to 100 notifications into one request, and the Ack
// applies to the whole POST — a consumer that fails on one record cannot ack the rest,
// so all of them are redelivered. Downstream dedupe has to account for that.
func TestParseOutboundMessageExpandsBatchedNotifications(t *testing.T) {
	t.Parallel()

	events := parseFixtureEvents(t, fixtureBatchUpdateContacts)

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	wantIDs := []string{"003xx000004TmiQAAS", "003xx000004TmiRAAS", "003xx000004TmiSAAS"}

	for index, wantID := range wantIDs {
		recordID, err := events[index].RecordId()
		if err != nil || recordID != wantID {
			t.Errorf("RecordId(%d) = %q, %v; want %q", index, recordID, err, wantID)
		}

		objectName, err := events[index].ObjectName()
		if err != nil || objectName != "Contact" {
			t.Errorf("ObjectName(%d) = %q, %v; want Contact", index, objectName, err)
		}

		// Every record in the batch predates the edit, so all three are updates.
		eventType, err := events[index].EventType()
		if err != nil || eventType != common.SubscriptionEventTypeUpdate {
			t.Errorf("EventType(%d) = %q, %v; want update", index, eventType, err)
		}

		// One bulk edit, so every notification carries the same LastModifiedDate:
		// 2026-09-09T00:41:57.000Z.
		nano, err := events[index].EventTimeStampNano()
		if err != nil || nano != int64(1788914517000000000) {
			t.Errorf("EventTimeStampNano(%d) = %d, %v; want 1788914517000000000", index, nano, err)
		}
	}
}

// TestOutboundMessageOmitsNullFields pins how outbound messages represent a null
// field, which is not what the SOAP conventions elsewhere in the envelope suggest.
//
// Both fixtures are the same Account under the same outbound message — deployed with
// Fields: ["Industry"], so Id, CreatedDate, LastModifiedDate and Industry were all
// requested. The only difference is that Industry was empty at create and set to
// "Agriculture" by the update.
//
// Salesforce does not send xsi:nil for a null sObject field. It omits the element
// entirely. (The one xsi:nil in the envelope, SessionId, sits on <notifications>,
// outside <sObject>, so it never reaches the field decoder.)
//
// The consequence for consumers: an absent key does not mean "not requested", it
// means "null". Absent and null are indistinguishable from the payload alone — only
// the deploy-time field list separates them.
func TestOutboundMessageOmitsNullFields(t *testing.T) {
	t.Parallel()

	sObjectFields := func(fixture string) map[string]any {
		t.Helper()

		events := parseFixtureEvents(t, fixture)

		raw, err := events[0].RawMap()
		if err != nil {
			t.Fatalf("RawMap: %v", err)
		}

		fields, ok := raw[omKeySObject].(map[string]any)
		if !ok {
			t.Fatalf("sObject missing from raw map: %v", raw)
		}

		return fields
	}

	populated := sObjectFields(fixtureFieldPopulated)
	if populated["Industry"] != "Agriculture" {
		t.Errorf("populated Industry = %v, want Agriculture", populated["Industry"])
	}

	omitted := sObjectFields(fixtureNullFieldOmitted)
	if value, exists := omitted["Industry"]; exists {
		t.Errorf("null Industry present as %v; Salesforce omits null fields entirely", value)
	}

	// The always-included fields survive in both, so an omitted Industry is the
	// null representation rather than a truncated payload.
	for name, fields := range map[string]map[string]any{
		"populated": populated, "omitted": omitted,
	} {
		for _, required := range []string{"Id", "CreatedDate", "LastModifiedDate"} {
			if _, exists := fields[required]; !exists {
				t.Errorf("%s payload missing guaranteed field %s", name, required)
			}
		}
	}
}
