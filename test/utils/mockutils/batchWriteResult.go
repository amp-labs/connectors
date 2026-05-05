package mockutils

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// BatchWriteResultComparator provides utility methods for comparing BatchWriteResult structures in tests.
//
// Unlike reflect.DeepEqual, these comparators support flexible (subset-based)
// data matching, allowing assertions on only relevant fields.
var BatchWriteResultComparator = batchWriteResultComparator{}

type batchWriteResultComparator struct{}

// SubsetWriteResults compares two BatchWriteResult objects and returns a CompareResult
// if each WriteResult in `expected` matches its corresponding entry in `actual`.
//
// A match is defined as follows:
//   - Subset equality for the Data field of each WriteResult (only expected keys/values are checked).
//   - Normalized equality for Errors, supporting struct/JSON, string, or golang error comparison.
//   - Exact equality for the Success and RecordId fields.
func (batchWriteResultComparator) SubsetWriteResults(
	actual, expected *common.BatchWriteResult,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	if len(actual.Results) != len(expected.Results) {
		result.AddDiff("expected %d batch results, got %d", len(expected.Results), len(actual.Results))
		return result
	}

	// Compare each result using existing comparator
	for i := range len(actual.Results) {
		actualResult := &actual.Results[i]
		expectedResult := &expected.Results[i]

		dataComparison := WriteResultComparator.SubsetData(actualResult, expectedResult)
		errorComparison := ErrorNormalizedComparator.EachErrorEquals(actualResult.Errors, expectedResult.Errors)

		for _, diff := range dataComparison.Diff {
			result.AddDiff("Result[%d] %s", i, diff)
		}

		for _, diff := range errorComparison.Diff {
			result.AddDiff("Result[%d] %s", i, diff)
		}

		result.Assert(fmt.Sprintf("Result[%d].Success", i), expectedResult.Success, actualResult.Success)
		result.Assert(fmt.Sprintf("Result[%d].RecordId", i), expectedResult.RecordId, actualResult.RecordId)

	}

	return result
}
