package mockutils

import (
	"reflect"

	"github.com/amp-labs/connectors/common"
)

var (
	WriteResultComparator = writeResultComparator{}
)

type writeResultComparator struct{}

// SubsetData checks that expected WriteResult.Data is a subset of actual WriteResult.Data
// other fields are strictly compared.
func (writeResultComparator) SubsetData(actual, expected *common.WriteResult) bool {
	if len(actual.Data) < len(expected.Data) {
		return false
	}

	// strict comparison
	ok := actual.Success == expected.Success &&
		actual.RecordId == expected.RecordId &&
		reflect.DeepEqual(actual.Errors, expected.Errors)

	if !ok {
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
