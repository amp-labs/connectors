package core

import (
	"errors"

	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeSOQL is the raw filter type Salesforce supports:
// a SOQL condition, as it would appear after WHERE.
const RawFilterTypeSOQL = "soql"

var errRawFilterUnbalanced = errors.New("has unbalanced parentheses")

// SOQLFilter is Salesforce's raw filter: a SOQL condition, as it would appear after WHERE.
type SOQLFilter struct {
	Condition string
}

// RawFilterType returns RawFilterTypeSOQL.
func (SOQLFilter) RawFilterType() string {
	return RawFilterTypeSOQL
}

// FromFilter accepts a non-blank string condition with balanced parentheses.
// Salesforce checks the rest of the condition when the query runs.
func (SOQLFilter) FromFilter(filter any) (SOQLFilter, error) {
	condition, err := common.StringFilterValue(filter)
	if err != nil {
		return SOQLFilter{}, err
	}

	if !parenthesesBalanced(condition) {
		return SOQLFilter{}, errRawFilterUnbalanced
	}

	return SOQLFilter{Condition: condition}, nil
}

// parenthesesBalanced reports whether every parenthesis outside string literals is matched.
// The condition is wrapped in parentheses when it is added to a query, so an unmatched ")" would
// close them early and let the rest of the condition escape the read's time window.
// Salesforce would accept such a query, so this is the one syntax check done here.
// It is a single pass over the condition with one case per kind of character.
func parenthesesBalanced(condition string) bool { // nolint:cyclop
	var (
		depth     int
		inLiteral bool
		escaped   bool
	)

	for _, char := range condition {
		switch {
		case inLiteral && escaped:
			escaped = false
		case inLiteral && char == '\\':
			escaped = true
		case inLiteral && char == '\'':
			inLiteral = false
		case inLiteral:
		case char == '\'':
			inLiteral = true
		case char == '(':
			depth++
		case char == ')':
			depth--

			if depth < 0 {
				return false
			}
		}
	}

	return depth == 0
}
