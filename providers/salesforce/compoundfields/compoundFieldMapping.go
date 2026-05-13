package compoundfields

import "strings"

// FlattenedCompoundSubField returns the flattened SOQL column name for the
// given object's compound sub-component (e.g. ("Contact", "MailingAddress",
// "Street") -> "mailingstreet"). Lookups are case-insensitive. Returns ("",
// false) when the (object, compound, sub-field) combination is not present in
// compoundFieldFlattenedByObject (schema.go).
//
// Use this to translate Salesforce CDC's "<Compound>.<Sub>" dot notation in
// ChangeEventHeader.changedFields into the flattened column name that
// downstream consumers (selectedFieldMappings, requiredWatchFields, SOQL
// queries) reference. Returned names are lowercased to match schema.go.
//
// See https://ampersand.slab.com/posts/salesforce-compound-fields-73jhbsjm
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
