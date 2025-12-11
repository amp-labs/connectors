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
	// We are expecting more fields than there in the existence.
	if len(actual.Data) < len(expected.Data) {
		return false
	}

	// At least one field should be mentioned.
	if len(actual.Data) > 0 && len(expected.Data) == 0 {
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
