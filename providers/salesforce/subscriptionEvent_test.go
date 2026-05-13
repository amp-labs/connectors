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

	// "Name" is not an *Address compound, so "Name.FirstName" stays dot-notation.
	// See providers/salesforce/compoundfields.
	assert.Equal(t, len(fields), 3, "should have three updated fields")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "LastModifiedById", "second field name should be LastModifiedById")
	assert.Equal(t, fields[2], "Name.FirstName", "third field name should preserve compound dot notation")
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
	assert.Equal(t, fields[1], "Name.LastName", "second field name should preserve Name compound dot notation")
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
		"mailingstreet",
		"mailingcity",
		"otherpostalcode",
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
		"billingstreet",
		"shippingpostalcode",
	})
}

func TestFlattenedCompoundSubField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		object   string
		compound string
		sub      string
		want     string
	}{
		// Account.BillingAddress: documented at sforce_api_objects_account.htm.
		{object: "Account", compound: "BillingAddress", sub: "Street", want: "billingstreet"},
		{object: "Account", compound: "BillingAddress", sub: "City", want: "billingcity"},
		{object: "Account", compound: "BillingAddress", sub: "State", want: "billingstate"},
		{object: "Account", compound: "BillingAddress", sub: "StateCode", want: "billingstatecode"},
		{object: "Account", compound: "BillingAddress", sub: "PostalCode", want: "billingpostalcode"},
		{object: "Account", compound: "BillingAddress", sub: "Country", want: "billingcountry"},
		{object: "Account", compound: "BillingAddress", sub: "CountryCode", want: "billingcountrycode"},
		{object: "Account", compound: "BillingAddress", sub: "GeocodeAccuracy", want: "billinggeocodeaccuracy"},
		{object: "Account", compound: "ShippingAddress", sub: "City", want: "shippingcity"},

		// Contact.MailingAddress / OtherAddress / Name:
		// documented at sforce_api_objects_contact.htm.
		{object: "Contact", compound: "MailingAddress", sub: "Street", want: "mailingstreet"},
		{object: "Contact", compound: "MailingAddress", sub: "PostalCode", want: "mailingpostalcode"},
		{object: "Contact", compound: "OtherAddress", sub: "PostalCode", want: "otherpostalcode"},
		{object: "Contact", compound: "Name", sub: "FirstName", want: ""},
		{object: "Contact", compound: "Name", sub: "LastName", want: ""},
		{object: "Contact", compound: "Name", sub: "Salutation", want: ""},

		// Lead.Address (single Address compound, no prefix on flattened columns):
		// documented at sforce_api_objects_lead.htm.
		{object: "Lead", compound: "Address", sub: "Street", want: "street"},
		{object: "Lead", compound: "Address", sub: "PostalCode", want: "postalcode"},
		{object: "Lead", compound: "Address", sub: "GeocodeAccuracy", want: "geocodeaccuracy"},
		{object: "Lead", compound: "Name", sub: "LastName", want: ""},

		// Case-insensitive compound / sub-field matching.
		{object: "contact", compound: "mailingaddress", sub: "street", want: "mailingstreet"},
		{object: "CONTACT", compound: "MAILINGADDRESS", sub: "STREET", want: "mailingstreet"},

		// *Address flattening applies for any object API name.
		{object: "AnotherObject", compound: "MailingAddress", sub: "Street", want: "mailingstreet"},
		{object: "AnotherObject", compound: "BillingAddress", sub: "City", want: "billingcity"},
		{object: "CustomObject__c", compound: "MailingAddress", sub: "Street", want: "mailingstreet"},

		// Opportunity has no compound fields per the API reference.
		{object: "Opportunity", compound: "Fiscal", sub: "Quarter", want: ""},

		// User + Name: not an address compound.
		{object: "User", compound: "Name", sub: "FirstName", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.object+"."+tc.compound+"."+tc.sub, func(t *testing.T) {
			t.Parallel()

			got := FlattenedCompoundSubField(tc.compound, tc.sub)
			assert.Equal(t, got, tc.want)
		})
	}
}
