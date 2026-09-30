package core

import "github.com/amp-labs/connectors/common"

// RawFilterTypeSOQL is the raw filter type Salesforce supports:
// a SOQL condition, as it would appear after WHERE.
const RawFilterTypeSOQL = "soql"

// SOQLFilter is Salesforce's raw filter: a SOQL condition, as it would appear after WHERE.
type SOQLFilter struct {
	Condition string
}

// RawFilterType returns RawFilterTypeSOQL.
func (SOQLFilter) RawFilterType() string {
	return RawFilterTypeSOQL
}

// FromFilter accepts a non-blank string condition. Salesforce checks the condition itself when the query runs.
func (SOQLFilter) FromFilter(filter any) (SOQLFilter, error) {
	condition, err := common.StringFilterValue(filter)
	if err != nil {
		return SOQLFilter{}, err
	}

	return SOQLFilter{Condition: condition}, nil
}
