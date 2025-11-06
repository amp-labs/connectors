package mockutils

import (
	"github.com/amp-labs/connectors/common"
)

// BatchWriteResultComparator provides utility methods for comparing BatchWriteResult structures in tests.
//
// Unlike reflect.DeepEqual, these comparators support flexible (subset-based)
// data matching, allowing assertions on only relevant fields.
var BatchWriteResultComparator = batchWriteResultComparator{}

type batchWriteResultComparator struct{}

// SubsetWriteResults compares two BatchWriteResult objects and returns true
// if each WriteResult in `expected` matches its corresponding entry in `actual`.
//
// A match is defined as follows:
//   - Subset equality for the Data field of each WriteResult (only expected keys/values are checked).
//   - Normalized equality for Errors, supporting struct/JSON, string, or golang error comparison.
//   - Exact equality for the Success and RecordId fields.
func (batchWriteResultComparator) SubsetWriteResults(actual, expected *common.BatchWriteResult) bool {
	if len(actual.Results) != len(expected.Results) {
		return false
	}

	// Compare each result using existing comparator
	for i := range len(actual.Results) {
		actualResult := &actual.Results[i]
		expectedResult := &expected.Results[i]

		a := WriteResultComparator.SubsetData(actualResult, expectedResult)
		b := ErrorNormalizedComparator.EachErrorEquals(actualResult.Errors, expectedResult.Errors)
		c := actualResult.Success == expectedResult.Success &&
			actualResult.RecordId == expectedResult.RecordId

		if !(a && b && c) {
			return false
		}
	}

	return true
}
