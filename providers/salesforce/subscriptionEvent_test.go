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

	// User has no rows in the compound field schema, so "Name.FirstName" is passed through
	// as the original dot notation rather than reduced to "firstname".
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

	assert.Equal(t, len(fields), 2, "should have one updated field")
	assert.Equal(t, fields[0], "LastModifiedDate", "first field name should be LastModifiedDate")
	assert.Equal(t, fields[1], "lastname", "second field name should be lastname")
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
		wantOK   bool
	}{
		// Account.BillingAddress: documented at sforce_api_objects_account.htm.
		{object: "Account", compound: "BillingAddress", sub: "Street", want: "billingstreet", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "City", want: "billingcity", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "State", want: "billingstate", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "StateCode", want: "billingstatecode", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "PostalCode", want: "billingpostalcode", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "Country", want: "billingcountry", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "CountryCode", want: "billingcountrycode", wantOK: true},
		{object: "Account", compound: "BillingAddress", sub: "GeocodeAccuracy", want: "billinggeocodeaccuracy", wantOK: true},
		{object: "Account", compound: "ShippingAddress", sub: "City", want: "shippingcity", wantOK: true},

		// Contact.MailingAddress / OtherAddress / Name:
		// documented at sforce_api_objects_contact.htm.
		{object: "Contact", compound: "MailingAddress", sub: "Street", want: "mailingstreet", wantOK: true},
		{object: "Contact", compound: "MailingAddress", sub: "PostalCode", want: "mailingpostalcode", wantOK: true},
		{object: "Contact", compound: "OtherAddress", sub: "PostalCode", want: "otherpostalcode", wantOK: true},
		{object: "Contact", compound: "Name", sub: "FirstName", want: "firstname", wantOK: true},
		{object: "Contact", compound: "Name", sub: "LastName", want: "lastname", wantOK: true},
		{object: "Contact", compound: "Name", sub: "Salutation", want: "salutation", wantOK: true},

		// Lead.Address (single Address compound, no prefix on flattened columns):
		// documented at sforce_api_objects_lead.htm.
		{object: "Lead", compound: "Address", sub: "Street", want: "street", wantOK: true},
		{object: "Lead", compound: "Address", sub: "PostalCode", want: "postalcode", wantOK: true},
		{object: "Lead", compound: "Address", sub: "GeocodeAccuracy", want: "geocodeaccuracy", wantOK: true},
		{object: "Lead", compound: "Name", sub: "LastName", want: "lastname", wantOK: true},

		// Case-insensitive lookups.
		{object: "contact", compound: "mailingaddress", sub: "street", want: "mailingstreet", wantOK: true},
		{object: "CONTACT", compound: "MAILINGADDRESS", sub: "STREET", want: "mailingstreet", wantOK: true},

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
