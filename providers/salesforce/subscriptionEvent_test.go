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

	// User is not in compoundFieldMappings, so "Name.FirstName" is passed through
	// as the original dot notation rather than reduced to "FirstName".
	// See compoundFieldMapping.go.
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

func TestFlattenedCompoundSubField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		object   string
		compound string
		sub      string
		want     string
		wantOK   bool
	}{
		// Account.BillingAddress: documented at sforce_api_objects_account.htm.
		{object: "Account", compound: "BillingAddress", sub: "Street", want: "BillingStreet", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "City", want: "BillingCity", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "State", want: "BillingState", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "StateCode", want: "BillingStateCode", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "PostalCode", want: "BillingPostalCode", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "Country", want: "BillingCountry", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "CountryCode", want: "BillingCountryCode", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "GeocodeAccuracy", want: "BillingGeocodeAccuracy", wantOK: true},
		{object: "Account", compound: "ShippingAddress", sub: "City", want: "ShippingCity", wantOK: true},

		// Contact.MailingAddress / OtherAddress / Name:
		// documented at sforce_api_objects_contact.htm.
		{object: "Contact", compound: "MailingAddress", sub: "Street", want: "MailingStreet", wantOK: true},
		{object: "Contact", compound: "MailingAddress", sub: "PostalCode", want: "MailingPostalCode", wantOK: true},
		{object: "Contact", compound: "OtherAddress", sub: "PostalCode", want: "OtherPostalCode", wantOK: true},
		{object: "Contact", compound: "Name", sub: "FirstName", want: "FirstName", wantOK: true},
		{object: "Contact", compound: "Name", sub: "LastName", want: "LastName", wantOK: true},
		{object: "Contact", compound: "Name", sub: "Salutation", want: "Salutation", wantOK: true},

		// Lead.Address (single Address compound, no prefix on flattened columns):
		// documented at sforce_api_objects_lead.htm.
		{object: "Lead", compound: "Address", sub: "Street", want: "Street", wantOK: true},
		{object: "Lead", compound: "Address", sub: "PostalCode", want: "PostalCode", wantOK: true},
		{object: "Lead", compound: "Address", sub: "GeocodeAccuracy", want: "GeocodeAccuracy", wantOK: true},
		{object: "Lead", compound: "Name", sub: "LastName", want: "LastName", wantOK: true},

		// Case-insensitive lookups.
		{object: "contact", compound: "mailingaddress", sub: "street", want: "MailingStreet", wantOK: true},
		{object: "CONTACT", compound: "MAILINGADDRESS", sub: "STREET", want: "MailingStreet", wantOK: true},

		// Opportunity has no compound fields per the API reference.
		{object: "Opportunity", compound: "Fiscal", sub: "Quarter", wantOK: false},

		// Unmapped objects fall through.
		{object: "User", compound: "Name", sub: "FirstName", wantOK: false},
		{object: "CustomObject__c", compound: "MailingAddress", sub: "Street", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.object+"."+tc.compound+"."+tc.sub, func(t *testing.T) {
			t.Parallel()

			got, ok := FlattenedCompoundSubField(tc.object, tc.compound, tc.sub)
			assert.Equal(t, ok, tc.wantOK)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestCompoundFieldFromFlattened(t *testing.T) {
	t.Parallel()

	tests := []struct {
		object       string
		flattened    string
		wantCompound string
		wantSub      string
		wantOK       bool
	}{
		{object: "Account", flattened: "BillingStreet", wantCompound: "BillingAddress", wantSub: "Street", wantOK: true},
		{object: "Account", flattened: "ShippingPostalCode", wantCompound: "ShippingAddress", wantSub: "PostalCode", wantOK: true},
		{object: "Contact", flattened: "MailingCity", wantCompound: "MailingAddress", wantSub: "City", wantOK: true},
		{object: "Contact", flattened: "OtherStreet", wantCompound: "OtherAddress", wantSub: "Street", wantOK: true},
		{object: "Contact", flattened: "FirstName", wantCompound: "Name", wantSub: "FirstName", wantOK: true},
		{object: "Lead", flattened: "Street", wantCompound: "Address", wantSub: "Street", wantOK: true},
		{object: "Lead", flattened: "LastName", wantCompound: "Name", wantSub: "LastName", wantOK: true},

		// Case-insensitive.
		{object: "contact", flattened: "mailingstreet", wantCompound: "MailingAddress", wantSub: "Street", wantOK: true},

		// Not a compound sub-field on Account.
		{object: "Account", flattened: "Industry", wantOK: false},
		// Not a mapped object.
		{object: "User", flattened: "FirstName", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.object+"."+tc.flattened, func(t *testing.T) {
			t.Parallel()

			compound, sub, ok := CompoundFieldFromFlattened(tc.object, tc.flattened)
			assert.Equal(t, ok, tc.wantOK)
			assert.Equal(t, compound, tc.wantCompound)
			assert.Equal(t, sub, tc.wantSub)
		})
	}
}
