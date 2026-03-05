package testroutines

import (
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

// URLTestServer is an alias to mock server BaseURL.
// For usage please refer to ComparatorPagination.
const URLTestServer = "{{testServerURL}}"

// Comparator is an equality function with custom rules.
// This package provides the most commonly used comparators.
type Comparator[Output any] func(serverURL string, actual, expected Output) bool

// ComparatorSubsetRead ensures that a subset of fields or raw data is present in the response.
// This is convenient for cases where the returned data is large,
// allowing for a more concise test that still validates the desired behavior.
func ComparatorSubsetRead(serverURL string, actual, expected *common.ReadResult) bool {
	a := mockutils.ReadResultComparator.SubsetFields(actual, expected)
	b := mockutils.ReadResultComparator.SubsetRaw(actual, expected)
	c := ComparatorPagination(serverURL, actual, expected)

	return a &&
		b &&
		c
}

// ComparatorPagination will check pagination related fields.
// Note: you may use an alias for Mock-Server-URL which will be dynamically resolved at runtime.
// Example:
//
//		common.ReadResult{
//			NextPage: testroutines.URLTestServer + "/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTI="
//	 }
//
// At runtime this may look as follows: http://127.0.0.1:38653/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTI=.
// The query parameters in URL can be in different order, encoding could differ as soon as the URL content matches
// the check will conclude that pagination matches.
func ComparatorPagination(serverURL string, actual *common.ReadResult, expected *common.ReadResult) bool {
	expectedNextPage := resolveTestServerURL(expected.NextPage.String(), serverURL)

	a := compareNextPageToken(actual.NextPage.String(), expectedNextPage)
	b := actual.Rows == expected.Rows
	c := actual.Done == expected.Done

	return a &&
		b &&
		c
}

func compareNextPageToken(actual, expected string) bool {
	if len(actual) == 0 && len(expected) == 0 {
		return true
	}

	if !strings.HasPrefix(actual, "http") {
		// Next page token is not a URL, compare raw text.
		return actual == expected
	}

	// We are dealing with URLs.
	// Compare URLs ignoring the query parameter order or encoding.
	// However, the "data content" must match.
	actualURL, err := urlbuilder.New(actual)
	if err != nil {
		return false
	}

	expectedURL, err := urlbuilder.New(expected)
	if err != nil {
		return false
	}

	return actualURL.Equals(expectedURL)
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
func ComparatorSubsetWrite(_ string, actual, expected *common.WriteResult) bool {
	return mockutils.WriteResultComparator.SubsetData(actual, expected) &&
		mockutils.ErrorNormalizedComparator.EachErrorEquals(actual.Errors, expected.Errors) &&
		actual.Success == expected.Success &&
		actual.RecordId == expected.RecordId
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
func ComparatorSubsetBatchWrite(_ string, actual, expected *common.BatchWriteResult) bool {
	if actual.Status != expected.Status ||
		actual.SuccessCount != expected.SuccessCount ||
		actual.FailureCount != expected.FailureCount {
		return false
	}

	return mockutils.BatchWriteResultComparator.SubsetWriteResults(actual, expected) &&
		mockutils.ErrorNormalizedComparator.EachErrorEquals(actual.Errors, expected.Errors)
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
func ComparatorSubsetMetadata(_ string, actual, expected *common.ListObjectMetadataResult) bool {
	if len(expected.Result)+len(expected.Errors) == 0 {
		panic("please specify expected Result or Errors in Metadata response")
	}

	return mockutils.MetadataResultComparator.SubsetFields(actual, expected) &&
		mockutils.MetadataResultComparator.SubsetErrors(actual, expected)
}

func resolveTestServerURL(urlTemplate string, serverURL string) string {
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
func ComparatorSubsetUpsertMetadata(_ string, actual, expected *common.UpsertMetadataResult) bool {
	if actual.Success != expected.Success || len(actual.Fields) != len(expected.Fields) {
		return false
	}

	// Compare field results.
	// When first difference is found, return with failure.
	for propertyName, property := range expected.Fields {
		for fieldName, expectedField := range property {
			actualProperty, ok := actual.Fields[propertyName]
			if !ok {
				return false
			}

			actualField, ok := actualProperty[fieldName]
			if !ok {
				return false
			}

			// Field properties should be the same. This is a hard comparison.
			generalPropertiesIdentical := actualField.FieldName == expectedField.FieldName &&
				actualField.Action == expectedField.Action &&
				reflect.DeepEqual(actualField.Warnings, expectedField.Warnings)

			if !generalPropertiesIdentical {
				return false
			}

			// A set of expected fields must be present
			if !mapIsSubsetMap(expectedField.Metadata, actualField.Metadata) {
				return false
			}
		}
	}

	// Everything is the same.
	return true
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
