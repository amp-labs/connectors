package search

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// makeSOQL returns the SOQL query for the desired search operation.
// It fails if the params carry a raw filter that isn't a valid Salesforce raw filter.
func makeSOQL(params *common.SearchParams) (*core.SOQLBuilder, error) {
	fields := associations.FieldsForSelectQuerySearch(params)
	soql := (&core.SOQLBuilder{}).
		SelectFields(fields).
		From(params.ObjectName)

	if err := addWhereClauses(soql, params); err != nil {
		return nil, err
	}

	return soql, nil
}

// addWhereClauses adds WHERE clauses to the SOQL query based on the params.
// Search params carry either field filters or a raw filter, never both (SearchParams.ValidateParams).
func addWhereClauses(soql *core.SOQLBuilder, params *common.SearchParams) error {
	if params.RawFilter != nil {
		filter, err := common.RawFilterAs[core.SOQLFilter](*params.RawFilter)
		if err != nil {
			return err
		}

		soql.WhereRawFilter(filter.Condition)

		return nil
	}

	// nolint:lll
	// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_comparisonoperators.htm
	for _, filter := range params.Filter.FieldFilters {
		if filter.Operator == common.FilterOperatorEQ {
			if isNumeric(filter.Value) {
				// No quotes for numbers.
				soql.Where(fmt.Sprintf("%s = %v", filter.FieldName, filter.Value))
			} else {
				// With quotes for all other values.
				soql.Where(fmt.Sprintf("%s = '%v'", filter.FieldName, filter.Value))
			}
		}
	}

	return nil
}

func isNumeric(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}
