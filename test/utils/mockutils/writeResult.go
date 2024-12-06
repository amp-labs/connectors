package mockutils

import (
	"reflect"

	"github.com/amp-labs/connectors/common"
)

var WriteResultComparator = writeResultComparator{}

type writeResultComparator struct{}

// SubsetData checks that expected WriteResult.Data is a subset of actual WriteResult.Data
// other fields are strictly compared.
func (writeResultComparator) SubsetData(actual, expected *common.WriteResult) bool {
	if len(actual.Data) < len(expected.Data) {
		return false
	}

	for k, expectedValue := range expected.Data {
		actualValue, ok := actual.Data[k]
		if !ok {
			return false
		}

		if !reflect.DeepEqual(actualValue, expectedValue) {
			return false
		}
	}

	return true
}

// ExactErrors uses strict error comparison.
func (writeResultComparator) ExactErrors(actual, expected *common.WriteResult) bool {
	return reflect.DeepEqual(actual.Errors, expected.Errors)
}
