package testconn

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// URLTestServer is an alias to mock server BaseURL.
// For usage please refer to ComparatorPagination.
const URLTestServer = "{{testServerURL}}"

// Comparator is an equality function with custom rules for specific test scenarios.
// Takes server URL actual output, expected output, and returns detailed comparison result.
//
// This package provides the most commonly used comparators like ComparatorSubsetRead, ComparatorPagination,
// ComparatorSubsetWrite, ComparatorSubsetMetadata for partial field matching in large API responses.
type Comparator[Output any] func(serverURL string, actual, expected Output) *testutils.CompareResult

// ComparatorSubsetRead ensures that a subset of fields or raw data is present in the response.
// This is convenient for cases where the returned data is large,
// allowing for a more concise test that still validates the desired behavior.
func ComparatorSubsetRead(serverURL string, actual, expected *common.ReadResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	if actual == nil && expected == nil {
		return result
	}

	if actual == nil {
		result.AddDiff("actual ReadResult is empty while expected something")
		return result
	}

	if expected == nil {
		result.AddDiff("expected ReadResult should not be empty")
		return result
	}

	result.Merge(mockutils.ReadResultComparator.SubsetFields(actual, expected))
	result.Merge(mockutils.ReadResultComparator.SubsetRaw(actual, expected))
	result.Merge(mockutils.ReadResultComparator.SubsetAssociationsRaw(actual, expected))
	result.Merge(mockutils.ReadResultComparator.Identifiers(actual, expected))
	result.Merge(ComparatorPagination(serverURL, actual, expected))

	return result
}

// ComparatorSubsetReadSorted is similar to ComparatorSubsetRead but the actual Rows are sorted using the identifiers.
// This ensures that the rows returned by connector are in the same order for the testing purposes.
// The test expectation should follow this imposed order.
// This is important to preserve the indexes of the test reports.
//
// Sort: Ascending order of common.ReadResultRow.Id.
func ComparatorSubsetReadSorted(serverURL string, actual, expected *common.ReadResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	if actual == nil && expected == nil {
		return result
	}

	if actual == nil {
		result.AddDiff("actual ReadResult is empty while expected something")
		return result
	}

	if expected == nil {
		result.AddDiff("expected ReadResult should not be empty")
		return result
	}

	sort.Slice(actual.Data, func(i, j int) bool {
		return actual.Data[i].Id < actual.Data[j].Id
	})

	return ComparatorSubsetRead(serverURL, actual, expected)
}

// ComparatorSortedSubsetReadByIds compares two slices of ReadResultRow as a subset,
// ignoring order and focusing only on relevant fields: raw data, associations, and identifiers.
//
// The `actual` slice is sorted by ID to ensure consistent output.
// The `expected` slice must be pre-sorted in the desired order. We intentionally do not sort
// `expected` so that mismatch logs can correctly report the index positions that differ.
//
// Sort: Ascending order of common.ReadResultRow.Id.
func ComparatorSortedSubsetReadByIds(serverURL string,
	actual, expected []common.ReadResultRow,
) *testutils.CompareResult {
	return ComparatorSubsetReadSorted(serverURL,
		&common.ReadResult{
			Rows: int64(len(actual)),
			Data: actual,
		},
		&common.ReadResult{
			Rows: int64(len(expected)),
			Data: expected,
		},
	)
}

// ComparatorPagination will check pagination related fields.
// Note: you may use an alias for Mock-Server-URL which will be dynamically resolved at runtime.
// Example:
//
//		common.ReadResult{
//			NextPage: testconn.URLTestServer + "/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTI="
//	 }
//
// At runtime this may look as follows: http://127.0.0.1:38653/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTI=.
// The query parameters in URL can be in different order, encoding could differ as soon as the URL content matches
// the check will conclude that pagination matches.
func ComparatorPagination(
	serverURL string, actual *common.ReadResult, expected *common.ReadResult,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	result.Assert("Rows", expected.Rows, actual.Rows)
	result.Assert("Done", expected.Done, actual.Done)

	expectedNextPage := ResolveTestServerURL(expected.NextPage.String(), serverURL)
	result.Merge(compareNextPageToken(actual.NextPage.String(), expectedNextPage))

	return result
}

func compareNextPageToken(actual, expected string) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	if actual == "" && expected == "" {
		return result
	}

	if actual == expected {
		// Raw text matches
		return result
	}

	if strings.HasPrefix(actual, "http") {
		// We are dealing with URLs.
		// Compare URLs ignoring the query parameter order or encoding.
		// However, the "data content" must match.
		result.Merge(compareHTTPURLs(expected, actual, nil))

		return result
	}

	// The token could be an aggregate token, compare the JSON format.
	return aggregateTokensMatch(expected, actual)
}

// ComparatorSubsetWrite compares two WriteResult objects, allowing partial
// (subset) matching for Data fields while requiring exact matches for Success and RecordId.
//
// It provides flexible error comparison logic:
//   - Errors are normalized before comparison, allowing strings, Go error types,
//     and mockutils.JSONErrorWrapper values (for JSON-based or struct comparison)
//     to be treated uniformly.
//
// This comparator is typically used when only a subset of Data fields
// needs verification rather than a full equality check.
func ComparatorSubsetWrite(_ string, actual, expected *common.WriteResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	result.Assert("Success", expected.Success, actual.Success)
	result.Assert("RecordId", expected.RecordId, actual.RecordId)
	result.Merge(mockutils.WriteResultComparator.SubsetData(actual, expected))
	result.Merge(mockutils.ErrorNormalizedComparator.EachErrorEquals(actual.Errors, expected.Errors))

	return result
}

// ComparatorSubsetBatchWrite compares two BatchWriteResult objects,
// performing subset matching for individual WriteResult entries while
// ensuring batch-level metrics (Status, SuccessCount, FailureCount) match exactly.
//
// Error comparison is normalized, allowing flexible matches between
// strings, Go errors, and mockutils.JSONErrorWrapper values—useful when
// top-level or per-record errors are represented as structs or JSON.
//
// This enables expressive, stable tests that verify meaningful fields
// without enforcing strict structural equality across the entire batch.
func ComparatorSubsetBatchWrite(_ string, actual, expected *common.BatchWriteResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	result.Assert("Status", expected.Status, actual.Status)
	result.Assert("SuccessCount", expected.SuccessCount, actual.SuccessCount)
	result.Assert("FailureCount", expected.FailureCount, actual.FailureCount)
	result.Merge(mockutils.BatchWriteResultComparator.SubsetWriteResults(actual, expected))
	result.Merge(mockutils.ErrorNormalizedComparator.EachErrorEquals(actual.Errors, expected.Errors))

	return result
}

// ComparatorSubsetMetadata will check a subset of fields is present.
// Errors could be an exact match for each object or subset can be used as well.
// This must be done by specifying expected errors using mockutils.ExpectedSubsetErrors.
// Then errors.Is() will be applied for each error.
//
// For if this is the case refer to the example below:
//
//	Errors: map[string]error{
//		"arsenal": mockutils.ExpectedSubsetErrors{ 						// Is doing a subset match.
//			common.ErrCaller,
//			errors.New(string(unsupportedResponse)),
//		},
//		"arsenal": common.NewHTTPError(http.StatusBadRequest,		// Is doing exact match.
//			headers, body, fmt.Errorf("%w: %s", common.ErrCaller, string(unsupportedResponse))),
//	},
func ComparatorSubsetMetadata(_ string, actual, expected *common.ListObjectMetadataResult) *testutils.CompareResult {
	if len(expected.Result)+len(expected.Errors) == 0 {
		panic("please specify expected Result or Errors in Metadata response")
	}

	result := testutils.NewCompareResult()
	result.Merge(mockutils.MetadataResultComparator.SubsetFields(actual, expected))
	result.Merge(mockutils.MetadataResultComparator.SubsetErrors(actual, expected))

	return result
}

// ComparatorSubsetMetadataWithMissingFields returns a comparator that checks
// metadata using ComparatorSubsetMetadata and also verifies that selected
// fields are absent from the returned object metadata.
//
// Use missingFields to specify object names and the fields that must not be
// present in either Result[objectName].Fields.
// Result[objectName].FieldsMap are supported for backwards compatability.
func ComparatorSubsetMetadataWithMissingFields(
	missingFields map[string][]string,
) Comparator[*common.ListObjectMetadataResult] {
	return func(serverURL string, actual, expected *common.ListObjectMetadataResult) *testutils.CompareResult {
		result := testutils.NewCompareResult()

		result.Merge(ComparatorSubsetMetadata(serverURL, actual, expected))
		for objectName, fields := range missingFields {
			objectMetadata, ok := actual.Result[objectName]
			if !ok {
				continue
			}

			for _, field := range fields {
				if _, ok = objectMetadata.Fields[field]; ok {
					result.AddDiff("Result[%v].Fields[%v] exists, but should be missing", objectName, field)
				}

				if _, ok = objectMetadata.FieldsMap[field]; ok {
					result.AddDiff("Result[%v].FieldsMap[%v] exists, but should be missing", objectName, field)
				}
			}
		}

		return result
	}
}

func ResolveTestServerURL(urlTemplate string, serverURL string) string {
	return strings.ReplaceAll(urlTemplate, URLTestServer, serverURL)
}

// ComparatorSubsetUpsertMetadata compares two UpsertMetadataResult objects,
// ensuring structural equality for core result properties while allowing
// subset matching for metadata contents.
//
// Comparison rules:
//
//   - Success must match exactly between actual and expected.
//   - The number of top-level field groups must match.
//   - For every expected property and field:
//   - The property and field must exist in the actual result.
//   - FieldName and Action must match exactly.
//   - Warnings must match exactly (DeepEqual comparison).
//   - Metadata is compared using subset semantics — all key/value
//     pairs defined in expected.Metadata must be present in
//     actual.Metadata, but actual may contain additional entries.
func ComparatorSubsetUpsertMetadata(_ string, actual, expected *common.UpsertMetadataResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	result.Assert(fmt.Sprintf("Success mismatch"), expected.Success, actual.Success)
	result.Assert(fmt.Sprintf("Fields length mismatch"), len(expected.Fields), len(actual.Fields))

	// Compare field results.
	for propertyName, property := range expected.Fields {
		for fieldName, expectedField := range property {
			actualProperty, ok := actual.Fields[propertyName]
			if !ok {
				result.AddDiff("Fields[%s] missing", propertyName)
				continue
			}

			actualField, ok := actualProperty[fieldName]
			if !ok {
				result.AddDiff("Fields[%s][%s] missing", propertyName, fieldName)
				continue
			}

			// Field properties should be the same. This is a hard comparison.
			result.Assert(fmt.Sprintf("Fields[%s][%s].FieldName", propertyName, fieldName),
				expectedField.FieldName, actualField.FieldName)
			result.Assert(fmt.Sprintf("Fields[%s][%s].Action", propertyName, fieldName),
				expectedField.Action, actualField.Action)
			result.Assert(fmt.Sprintf("Fields[%s][%s].Warnings", propertyName, fieldName),
				expectedField.Warnings, actualField.Warnings)

			// A set of expected fields must be present
			if !mapIsSubsetMap(expectedField.Metadata, actualField.Metadata) {
				result.Assert(fmt.Sprintf("Fields[%s][%s].Metadata", propertyName, fieldName),
					expectedField.Metadata, actualField.Metadata)
			}
		}
	}

	return result
}

// ComparatorSubscriptionWithResult returns a comparator for subscription results
// that first compares the common SubscriptionResult fields and then compares the
// nested Result values with the provided resultComparator.
//
// The generic type parameter R represents the concrete Result payload type.
//
// Both expected.Result and actual.Result must be pointers. Otherwise, this is not a valid
// connector implementation.
// If Result is of no importance use ComparatorSubscriptionWithoutResult.
func ComparatorSubscriptionWithResult[R any](
	resultComparator func(expectedResult, actualResult *R) *testutils.CompareResult,
) Comparator[*common.SubscriptionResult] {
	return func(_ string, actual, expected *common.SubscriptionResult) *testutils.CompareResult {
		result := testutils.NewCompareResult()
		result.Merge(mockutils.SubscriptionResultComparator.CompareWithoutResultArg(actual, expected))

		// Connector must return `Result` as a pointer.
		// If that is the case we proceed with comparing these results.
		if result.Assert("expected.Result must be pointer", true, isPointer(expected.Result)) &&
			result.Assert("actual.Result must be pointer", true, isPointer(actual.Result)) {
			result.Merge(resultComparator(expected.Result.(*R), actual.Result.(*R)))
		}

		return result
	}
}

// ComparatorSubscriptionWithoutResult compares subscription results without inspecting
// the nested Result payload. It is useful when the Result field is irrelevant for the test case.
//
// If Result data structure should be compared use ComparatorSubscriptionWithResult.
func ComparatorSubscriptionWithoutResult(
	_ string, actual, expected *common.SubscriptionResult,
) *testutils.CompareResult {
	return mockutils.SubscriptionResultComparator.CompareWithoutResultArg(actual, expected)
}

func isPointer(v any) bool {
	if v == nil {
		return false
	}
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}

func mapIsSubsetMap(subset, superset map[string]any) bool {
	for key, expected := range subset {
		actual, ok := superset[key]
		if !ok {
			return false // missing key
		}

		if !reflect.DeepEqual(expected, actual) {
			return false // values not the same
		}
	}

	return true
}

func compareHTTPURLs(expected string, actual string, index *int) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	actualURL, err := urlbuilder.New(actual)
	if err != nil {
		if index == nil {
			result.AddDiff("NextPage actual cannot be parsed as URL")
		} else {
			result.AddDiff("NextPage[%v] actual cannot be parsed as URL", *index)
		}
		return result
	}

	expectedURL, err := urlbuilder.New(expected)
	if err != nil {
		if index == nil {
			result.AddDiff("NextPage expected cannot be parsed as URL")
		} else {
			result.AddDiff("NextPage[%v] expected cannot be parsed as URL", *index)
		}
		return result
	}

	if !actualURL.Equals(expectedURL) {
		if index == nil {
			result.AddDiff("NextPage URLs do not match actual(%v), expected(%v)", actualURL, expectedURL)
		} else {
			result.AddDiff("NextPage[%v] URLs do not match actual(%v), expected(%v)", *index, actualURL, expectedURL)
		}
	}

	return result
}

func aggregateTokensMatch(expected string, actual string) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	actualAgg := make(readhelper.AggregateNextPage[any], 0)
	if err := json.Unmarshal([]byte(actual), &actualAgg); err != nil {
		// It is not an aggregate.
		result.AddDiff("NextPage mismatch actual(%v) expected(%v)", actual, expected)
		return result
	}

	expectedAgg := make(readhelper.AggregateNextPage[any], 0)
	if err := json.Unmarshal([]byte(expected), &expectedAgg); err != nil {
		// It is not an aggregate.
		result.AddDiff("NextPage mismatch actual(%v) expected(%v)", actual, expected)
		return result
	}

	if !result.Assert("NextPage len(AggregateNextPage)", len(actualAgg), len(expectedAgg)) {
		return result
	}

	sort.Slice(actualAgg, func(i, j int) bool {
		return actualAgg[i].Value.String() < actualAgg[j].Value.String() ||
			fmt.Sprintf("%v", actualAgg[i].Context) < fmt.Sprintf("%v", actualAgg[j].Context)
	})

	for index, actualToken := range actualAgg {
		actualURL := actualToken.Value.String()
		expectedURL := expectedAgg[index].Value.String()

		if strings.HasPrefix(actualURL, "http") {
			result.Merge(compareHTTPURLs(expectedURL, actualURL, new(index)))
		}

		result.Assert(fmt.Sprintf("NextPage[%v].Context", index), expectedAgg, actualAgg)
	}

	return result
}
