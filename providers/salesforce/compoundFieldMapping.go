package salesforce

import "strings"

// compoundFieldMapping describes how a standard Salesforce compound field's
// sub-component corresponds to the flattened column name used in SOQL and the
// "<Compound>.<Sub>" dot notation used in CDC ChangeEventHeader.changedFields.
//
// Example: on the Contact object, the "MailingAddress" compound field's
// "Street" sub-component is exposed as the flattened column "MailingStreet";
// CDC reports updates to it as "MailingAddress.Street".
type compoundFieldMapping struct {
	Object    string // Salesforce object API name, e.g. "Account"
	Compound  string // Compound field name, e.g. "BillingAddress"
	SubField  string // Sub-component name, e.g. "Street"
	Flattened string // Flattened column name, e.g. "BillingStreet"
}

// compoundFieldMappings is the source-of-truth list of standard Salesforce
// compound field sub-components and their flattened column names. Each entry
// was taken directly from the Salesforce API Object Reference:
//
//	Account:     https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_account.htm
//	Contact:     https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_contact.htm
//	Lead:        https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_lead.htm
//	Opportunity: https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_opportunity.htm
//
// Opportunity has no entries: per the API reference, Fiscal, FiscalQuarter,
// and FiscalYear are three independent fields, not a compound with
// sub-components.
//
// Add new objects here as additional providers are onboarded; the lookup
// helpers below are built from this slice at init.
var compoundFieldMappings = []compoundFieldMapping{ //nolint:gochecknoglobals
	// ---------------------------------------------------------------------
	// Account
	// ---------------------------------------------------------------------
	// BillingAddress compound.
	{Object: "Account", Compound: "BillingAddress", SubField: "Street", Flattened: "BillingStreet"},
	{Object: "Account", Compound: "BillingAddress", SubField: "City", Flattened: "BillingCity"},
	{Object: "Account", Compound: "BillingAddress", SubField: "State", Flattened: "BillingState"},
	{Object: "Account", Compound: "BillingAddress", SubField: "StateCode", Flattened: "BillingStateCode"},
	{Object: "Account", Compound: "BillingAddress", SubField: "PostalCode", Flattened: "BillingPostalCode"},
	{Object: "Account", Compound: "BillingAddress", SubField: "Country", Flattened: "BillingCountry"},
	{Object: "Account", Compound: "BillingAddress", SubField: "CountryCode", Flattened: "BillingCountryCode"},
	{Object: "Account", Compound: "BillingAddress", SubField: "Latitude", Flattened: "BillingLatitude"},
	{Object: "Account", Compound: "BillingAddress", SubField: "Longitude", Flattened: "BillingLongitude"},
	{Object: "Account", Compound: "BillingAddress", SubField: "GeocodeAccuracy", Flattened: "BillingGeocodeAccuracy"},
	// ShippingAddress compound.
	{Object: "Account", Compound: "ShippingAddress", SubField: "Street", Flattened: "ShippingStreet"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "City", Flattened: "ShippingCity"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "State", Flattened: "ShippingState"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "StateCode", Flattened: "ShippingStateCode"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "PostalCode", Flattened: "ShippingPostalCode"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "Country", Flattened: "ShippingCountry"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "CountryCode", Flattened: "ShippingCountryCode"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "Latitude", Flattened: "ShippingLatitude"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "Longitude", Flattened: "ShippingLongitude"},
	{Object: "Account", Compound: "ShippingAddress", SubField: "GeocodeAccuracy", Flattened: "ShippingGeocodeAccuracy"},

	// ---------------------------------------------------------------------
	// Contact
	// ---------------------------------------------------------------------
	// MailingAddress compound.
	{Object: "Contact", Compound: "MailingAddress", SubField: "Street", Flattened: "MailingStreet"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "City", Flattened: "MailingCity"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "State", Flattened: "MailingState"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "StateCode", Flattened: "MailingStateCode"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "PostalCode", Flattened: "MailingPostalCode"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "Country", Flattened: "MailingCountry"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "CountryCode", Flattened: "MailingCountryCode"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "Latitude", Flattened: "MailingLatitude"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "Longitude", Flattened: "MailingLongitude"},
	{Object: "Contact", Compound: "MailingAddress", SubField: "GeocodeAccuracy", Flattened: "MailingGeocodeAccuracy"},
	// OtherAddress compound.
	{Object: "Contact", Compound: "OtherAddress", SubField: "Street", Flattened: "OtherStreet"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "City", Flattened: "OtherCity"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "State", Flattened: "OtherState"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "StateCode", Flattened: "OtherStateCode"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "PostalCode", Flattened: "OtherPostalCode"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "Country", Flattened: "OtherCountry"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "CountryCode", Flattened: "OtherCountryCode"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "Latitude", Flattened: "OtherLatitude"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "Longitude", Flattened: "OtherLongitude"},
	{Object: "Contact", Compound: "OtherAddress", SubField: "GeocodeAccuracy", Flattened: "OtherGeocodeAccuracy"},
	// Name compound.
	{Object: "Contact", Compound: "Name", SubField: "FirstName", Flattened: "FirstName"},
	{Object: "Contact", Compound: "Name", SubField: "MiddleName", Flattened: "MiddleName"},
	{Object: "Contact", Compound: "Name", SubField: "LastName", Flattened: "LastName"},
	{Object: "Contact", Compound: "Name", SubField: "Salutation", Flattened: "Salutation"},
	{Object: "Contact", Compound: "Name", SubField: "Suffix", Flattened: "Suffix"},

	// ---------------------------------------------------------------------
	// Lead
	// ---------------------------------------------------------------------
	// Address compound (no prefix on the flattened columns).
	{Object: "Lead", Compound: "Address", SubField: "Street", Flattened: "Street"},
	{Object: "Lead", Compound: "Address", SubField: "City", Flattened: "City"},
	{Object: "Lead", Compound: "Address", SubField: "State", Flattened: "State"},
	{Object: "Lead", Compound: "Address", SubField: "StateCode", Flattened: "StateCode"},
	{Object: "Lead", Compound: "Address", SubField: "PostalCode", Flattened: "PostalCode"},
	{Object: "Lead", Compound: "Address", SubField: "Country", Flattened: "Country"},
	{Object: "Lead", Compound: "Address", SubField: "CountryCode", Flattened: "CountryCode"},
	{Object: "Lead", Compound: "Address", SubField: "Latitude", Flattened: "Latitude"},
	{Object: "Lead", Compound: "Address", SubField: "Longitude", Flattened: "Longitude"},
	{Object: "Lead", Compound: "Address", SubField: "GeocodeAccuracy", Flattened: "GeocodeAccuracy"},
	// Name compound.
	{Object: "Lead", Compound: "Name", SubField: "FirstName", Flattened: "FirstName"},
	{Object: "Lead", Compound: "Name", SubField: "MiddleName", Flattened: "MiddleName"},
	{Object: "Lead", Compound: "Name", SubField: "LastName", Flattened: "LastName"},
	{Object: "Lead", Compound: "Name", SubField: "Salutation", Flattened: "Salutation"},
	{Object: "Lead", Compound: "Name", SubField: "Suffix", Flattened: "Suffix"},

	// ---------------------------------------------------------------------
	// Opportunity: no compound fields per the API reference. Fiscal,
	// FiscalQuarter, and FiscalYear are three independent fields.
	// ---------------------------------------------------------------------
}

// compoundFieldFlattenedByObject is built from compoundFieldMappings at init.
// Shape: object -> compound -> sub-field -> flattened column name.
// All map keys are lowercased so lookups are case-insensitive; values preserve
// the canonical PascalCase from the API reference.
var compoundFieldFlattenedByObject map[string]map[string]map[string]string //nolint:gochecknoglobals

// compoundFieldByFlattenedByObject is the inverse: object -> flattened ->
// (compound, sub-field). Also case-insensitive keys with canonical-case values.
// Exposed via CompoundFieldFromFlattened for future consumers that need to go
// from a SOQL column name back to the compound form (e.g. mapping a customer's
// selectedFieldMappings value to the dot-notation CDC emits).
var compoundFieldByFlattenedByObject map[string]map[string]compoundFieldMapping //nolint:gochecknoglobals

func init() {
	compoundFieldFlattenedByObject = make(map[string]map[string]map[string]string)
	compoundFieldByFlattenedByObject = make(map[string]map[string]compoundFieldMapping)

	for _, entry := range compoundFieldMappings {
		obj := strings.ToLower(entry.Object)
		compound := strings.ToLower(entry.Compound)
		sub := strings.ToLower(entry.SubField)
		flat := strings.ToLower(entry.Flattened)

		if _, ok := compoundFieldFlattenedByObject[obj]; !ok {
			compoundFieldFlattenedByObject[obj] = make(map[string]map[string]string)
		}

		if _, ok := compoundFieldFlattenedByObject[obj][compound]; !ok {
			compoundFieldFlattenedByObject[obj][compound] = make(map[string]string)
		}

		compoundFieldFlattenedByObject[obj][compound][sub] = entry.Flattened

		if _, ok := compoundFieldByFlattenedByObject[obj]; !ok {
			compoundFieldByFlattenedByObject[obj] = make(map[string]compoundFieldMapping)
		}

		compoundFieldByFlattenedByObject[obj][flat] = entry
	}
}

// FlattenedCompoundSubField returns the flattened SOQL column name for the
// given object's compound sub-component (e.g. ("Contact", "MailingAddress",
// "Street") -> "MailingStreet"). Lookups are case-insensitive. Returns ("",
// false) when the (object, compound, sub-field) combination is not present in
// compoundFieldMappings.
//
// Use this to translate Salesforce CDC's "<Compound>.<Sub>" dot notation in
// ChangeEventHeader.changedFields into the flattened column name that
// downstream consumers (selectedFieldMappings, requiredWatchFields, SOQL
// queries) reference.
func FlattenedCompoundSubField(object, compound, subField string) (string, bool) {
	compounds, ok := compoundFieldFlattenedByObject[strings.ToLower(object)]
	if !ok {
		return "", false
	}

	subs, ok := compounds[strings.ToLower(compound)]
	if !ok {
		return "", false
	}

	flat, ok := subs[strings.ToLower(subField)]

	return flat, ok
}

// CompoundFieldFromFlattened is the inverse of FlattenedCompoundSubField: given
// an object and a flattened SOQL column name (e.g. ("Contact",
// "MailingStreet")), it returns the compound field name and sub-field name
// ("MailingAddress", "Street"). Returns ("", "", false) when the column is not
// part of a known compound. Lookups are case-insensitive.
func CompoundFieldFromFlattened(object, flattened string) (compound, subField string, ok bool) {
	flatMap, found := compoundFieldByFlattenedByObject[strings.ToLower(object)]
	if !found {
		return "", "", false
	}

	entry, found := flatMap[strings.ToLower(flattened)]
	if !found {
		return "", "", false
	}

	return entry.Compound, entry.SubField, true
}
