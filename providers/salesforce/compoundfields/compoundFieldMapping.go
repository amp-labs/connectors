package compoundfields

import "strings"

// FlattenedFieldNameFromCompoundField returns the flattened field name
// of a field received in a CDC change event.
// See https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_subscribe_compound_fields.htm

// It addressed the following 2 types of compound fields:
// 1. Address compound fields:
//   - "MailingAddress", "Street" -> "MailingStreet"
//   - "Address", "Street" -> "Street"
//
// 2. Name fields:
//   - "Name", "FirstName" -> "FirstName"
//   - "Name", "LastName" -> "LastName"
//   - "Name", "Salutation" -> "Salutation"
//   - "Name", "Suffix" -> "Suffix"
//
// Please note that geolocation fields is not explicitly addressed, since they are
// included in address compound fields.
//   - "MailingAddress", "Latitude" -> "MailingLatitude"
//
// See https://ampersand.slab.com/posts/salesforce-compound-fields-73jhbsjm for terminology.
func FlattenedFieldNameFromCompoundField(compoundFieldName, subFieldName string) string {
	// 1. Handle address compound fields
	flatAddressField, ok := flattenedFromAddressCompound(compoundFieldName, subFieldName)
	if ok {
		return flatAddressField
	}

	// 2. For all other compound fields, return the sub-field name
	// e.g. ("Name", "FirstName") -> "FirstName"
	return subFieldName
}

func flattenedFromAddressCompound(compound, subField string) (string, bool) {
	prefix, ok := addressCompoundPrefix(compound)
	if !ok {
		return "", false
	}

	return prefix + subField, true
}

func addressCompoundPrefix(compound string) (prefix string, ok bool) {
	const suffix = "Address"

	if len(compound) < len(suffix) {
		return "", false
	}

	i := len(compound) - len(suffix)
	if !strings.EqualFold(compound[i:], suffix) {
		return "", false
	}

	return compound[:i], true
}
