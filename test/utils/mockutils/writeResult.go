package mockutils

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

var WriteResultComparator = writeResultComparator{}

type writeResultComparator struct{}

// SubsetData checks that expected WriteResult.Data is a subset of actual WriteResult.Data
// other fields are strictly compared.
func (writeResultComparator) SubsetData(actual, expected *common.WriteResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	// We are expecting more fields than there in the existence.
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff("expected at least %d data fields, got %d", len(expected.Data), len(actual.Data))
		return result
	}

	// At least one field should be mentioned.
	if len(actual.Data) > 0 && len(expected.Data) == 0 {
		result.AddDiff("expected some data fields, but none were specified in expected")
		return result
	}

	for key, expectedValue := range expected.Data {
		actualValue, ok := actual.Data[key]
		if !ok {
			result.AddDiff("Data[%s] missing", key)
			continue
		}

		result.Assert(fmt.Sprintf("Data[%s]", key), expectedValue, actualValue)
	}

	return result
}
