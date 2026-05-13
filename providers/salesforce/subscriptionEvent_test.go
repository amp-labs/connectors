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

	assert.Equal(t, len(fields), 3, "should have three updated fields")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "LastModifiedById", "second field name should be LastModifiedById")
	assert.Equal(t, fields[2], "FirstName", "compound Name.FirstName field should be flattened to FirstName")
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

	assert.Equal(t, len(fields), 2, "should have two updated fields")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "LastName", "compound Name.LastName field should be flattened to LastName")
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
// (e.g. "MailingAddress.Street" -> "mailingstreet"). This is what downstream
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
		"MailingStreet", // flattened from MailingAddress.Street
		"MailingCity", // flattened from MailingAddress.City
		"OtherPostalCode", // flattened from OtherAddress.PostalCode
	})
}

func TestSubscriptionEventLeadAccountCompoundAddress(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscription/update_lead_compound_address.json")

	event := SubscriptionEvent{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to start a test, cannot parse data; error (%v)", err)
	}

	objectName, err := event.ObjectName()
	assert.NilError(t, err, "error should be nil")
	assert.Equal(t, objectName, "Lead", "ObjectName should be Lead")

	fields, err := event.UpdatedFields()
	assert.NilError(t, err, "error should be nil")

	assert.DeepEqual(t, fields, []string{
		"Street", // flattened from Address.Street
		"PostalCode", // flattened from Address.PostalCode
	})
}

func TestNormalizeUpdatedFieldName(t *testing.T) {
	t.Parallel()

	var s SubscriptionEvent

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain field unchanged", in: "LastModifiedDate", want: "LastModifiedDate"},
		{name: "name without dot is unchanged", in: "Industry", want: "Industry"},
		{name: "empty string", in: "", want: ""},
		{name: "name compound first name", in: "Name.FirstName", want: "FirstName"},
		{name: "name compound last name", in: "Name.LastName", want: "LastName"},
		{name: "name compound salutation", in: "Name.Salutation", want: "Salutation"},
		{name: "mailing address street", in: "MailingAddress.Street", want: "MailingStreet"},
		{name: "mailing address postal code", in: "MailingAddress.PostalCode", want: "MailingPostalCode"},
		{name: "billing address city", in: "BillingAddress.City", want: "BillingCity"},
		{name: "shipping address state code", in: "ShippingAddress.StateCode", want: "ShippingStateCode"},
		{name: "other address country", in: "OtherAddress.Country", want: "OtherCountry"},
		{name: "bare address street lead style", in: "Address.Street", want: "Street"},
		{name: "bare address geocode", in: "Address.GeocodeAccuracy", want: "GeocodeAccuracy"},
		{name: "3-word address field", in: "DeliverToAddress.Street", want: "DeliverToStreet"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := s.normalizeUpdatedFieldName(tc.in)
			assert.Equal(t, got, tc.want)
		})
	}
}
