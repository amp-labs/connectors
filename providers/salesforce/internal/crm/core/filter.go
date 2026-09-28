package core

import "github.com/amp-labs/connectors/common"

// RawFilterTypeSOQL is the raw filter type Salesforce supports:
// a SOQL condition, as it would appear after WHERE.
const RawFilterTypeSOQL = "soql"

// ValidateRawFilter checks that a raw filter is of type RawFilterTypeSOQL with a string condition.
// Salesforce checks the condition itself when the query runs.
func ValidateRawFilter(filter common.RawFilter) error {
	_, err := common.StringRawFilter(filter, RawFilterTypeSOQL)

	return err
}
