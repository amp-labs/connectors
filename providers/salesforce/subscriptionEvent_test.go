package salesforce

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

func TestSubscriptionEventUpdateUser(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscription/update_user.json")

	event := SubscriptionEvent{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	eventType, err := event.EventType()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, eventType, common.SubscriptionEventTypeUpdate, "EventType should be Update")

	rawEventType, err := event.RawEventName()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, rawEventType, "UPDATE", "RawEventName should be UPDATE")

	objectName, err := event.ObjectName()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, objectName, "User", "ObjectName should be User")

	workspace, err := event.Workspace()

	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, workspace, "", "Workspace should be empty")

	fields, err := event.UpdatedFields()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, len(fields), 3, "should have one updated field")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "LastModifiedById", "second field name should be LastModifiedById")
	assert.Equal(t, fields[2], "FirstName", "third field name should be FirstName")
}

func TestSubscriptionEventUpdateContact(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscription/update_contact.json")

	event := SubscriptionEvent{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	eventType, err := event.EventType()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, eventType, common.SubscriptionEventTypeUpdate, "EventType should be Update")

	rawEventType, err := event.RawEventName()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, rawEventType, "UPDATE", "RawEventName should be UPDATE")

	objectName, err := event.ObjectName()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, objectName, "Contact", "ObjectName should be Contact")

	workspace, err := event.Workspace()

	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, workspace, "", "Workspace should be empty")

	fields, err := event.UpdatedFields()
	assert.NilError(t, err, "error should be nil")

	assert.Equal(t, len(fields), 2, "should have one updated field")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "LastName", "second field name should be LastName")
}

func TestSubscriptionEventProperties(t *testing.T) {
	t.Parallel()

	eventNewAccountData := testutils.DataFromFile(t, "subscription/new_account.json")

	changeEvent := CollapsedSubscriptionEvent{}
	if err := json.Unmarshal(eventNewAccountData, &changeEvent); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	events, err := changeEvent.SubscriptionEventList()
	assert.NilError(t, err, "error should be nil")

	if len(events) != 1 {
		t.Fatalf("failed to start a test, expected to have only one event")
	}

	event := events[0]

	eventType, err := event.EventType()
	validateSubEvent(t, err, eventType, common.SubscriptionEventTypeCreate, "EventType")

	rawEventType, err := event.RawEventName()
	validateSubEvent(t, err, rawEventType, "CREATE", "RawEventName")

	objectName, err := event.ObjectName()
	validateSubEvent(t, err, objectName, "Account", "ObjectName")

	workspace, err := event.Workspace()
	validateSubEvent(t, err, workspace, "", "Workspace")

	recordID, err := event.RecordId()
	validateSubEvent(t, err, recordID, "0015f00002J9YYEAA3", "RecordId")

	timestamp, err := event.EventTimeStampNano()
	validateSubEvent(t, err, timestamp, 1712693965000, "EventTimeStampNano")
}

func validateSubEvent[V any](t *testing.T, err error, actual, expected V, methodName string) {
	t.Helper()

	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, actual, expected, "method "+methodName)
}

// TestSubscriptionEventUpdateContactCompoundAddress verifies that Salesforce CDC
// changedFields entries using "<Compound>.<Sub>" dot notation for address
// compounds are normalized to their flattened column name
// (e.g. "MailingAddress.Street" -> "MailingStreet"). This is what downstream
// subscriptions match against. See ENG-3919.
func TestSubscriptionEventUpdateContactCompoundAddress(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscription/update_contact_compound_address.json")

	event := SubscriptionEvent{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	objectName, err := event.ObjectName()
	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, objectName, "Contact", "ObjectName should be Contact")

	fields, err := event.UpdatedFields()
	assert.NilError(t, err, "error should be nil")

	assert.DeepEqual(t, fields, []string{
		"LastModifiedDate",
		"MailingStreet",
		"MailingCity",
		"OtherPostalCode",
	})
}

// TestSubscriptionEventUpdateAccountCompoundAddress verifies the same flattening
// for Account, which has both BillingAddress and ShippingAddress compounds.
func TestSubscriptionEventUpdateAccountCompoundAddress(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscription/update_account_compound_address.json")

	event := SubscriptionEvent{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	objectName, err := event.ObjectName()
	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, objectName, "Account", "ObjectName should be Account")

	fields, err := event.UpdatedFields()
	assert.NilError(t, err, "error should be nil")

	assert.DeepEqual(t, fields, []string{
		"BillingStreet",
		"ShippingPostalCode",
	})
}

func TestFlattenedCompoundFieldName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		compound string
		sub      string
		want     string
	}{
		// Address-typed compounds: strip "Address" suffix and prepend.
		{compound: "MailingAddress", sub: "Street", want: "MailingStreet"},
		{compound: "MailingAddress", sub: "City", want: "MailingCity"},
		{compound: "MailingAddress", sub: "State", want: "MailingState"},
		{compound: "MailingAddress", sub: "PostalCode", want: "MailingPostalCode"},
		{compound: "MailingAddress", sub: "Country", want: "MailingCountry"},
		{compound: "MailingAddress", sub: "CountryCode", want: "MailingCountryCode"},
		{compound: "BillingAddress", sub: "Street", want: "BillingStreet"},
		{compound: "ShippingAddress", sub: "City", want: "ShippingCity"},
		{compound: "OtherAddress", sub: "PostalCode", want: "OtherPostalCode"},
		// Compounds named just "Address" flatten to the bare sub-field.
		{compound: "Address", sub: "Street", want: "Street"},
		// Name compounds drop the prefix.
		{compound: "Name", sub: "FirstName", want: "FirstName"},
		{compound: "Name", sub: "LastName", want: "LastName"},
		// Other compound types keep the compound name as the prefix.
		{compound: "Fiscal", sub: "Quarter", want: "FiscalQuarter"},
		{compound: "Location", sub: "Latitude", want: "LocationLatitude"},
	}

	for _, tc := range tests {
		t.Run(tc.compound+"."+tc.sub, func(t *testing.T) {
			t.Parallel()

			got := flattenedCompoundFieldName(tc.compound, tc.sub)
			assert.Equal(t, got, tc.want)
		})
	}
}
