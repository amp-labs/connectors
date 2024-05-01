package mockutils

import (
	"fmt"
	"reflect"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
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

	for k, v := range expected.Data {
		actualValue, ok := actual.Data[k]
		if !ok {
			return false
		}

		if convertToString(actualValue) != convertToString(v) {
			return false
		}
	}

	return true
}

func convertToString(obj any) string {
	if node, ok := obj.(*ajson.Node); ok {
		val, err := node.Value()
		if err == nil {
			return fmt.Sprintf("%v", val)
		}
	}

	return fmt.Sprintf("%v", obj)
}
