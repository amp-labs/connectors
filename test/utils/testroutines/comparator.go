package testroutines

import (
	"strings"

	"github.com/amp-labs/connectors/common"
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
	return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
		mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
		ComparatorPagination(serverURL, actual, expected)
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
func ComparatorPagination(serverURL string, actual *common.ReadResult, expected *common.ReadResult) bool {
	expectedNextPage := resolveTestServerURL(expected.NextPage.String(), serverURL)

	return actual.NextPage.String() == expectedNextPage &&
		actual.Rows == expected.Rows &&
		actual.Done == expected.Done
}

// ComparatorSubsetWrite ensures that only the specified metadata objects are present,
// while other values are verified through an exact match..
func ComparatorSubsetWrite(_ string, actual, expected *common.WriteResult) bool {
	return mockutils.WriteResultComparator.SubsetData(actual, expected) &&
		mockutils.WriteResultComparator.ExactErrors(actual, expected) &&
		actual.Success == expected.Success &&
		actual.RecordId == expected.RecordId
}

// ComparatorSubsetMetadata will check a subset of fields is present.
func ComparatorSubsetMetadata(_ string, actual, expected *common.ListObjectMetadataResult) bool {
	return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
}

func resolveTestServerURL(urlTemplate string, serverURL string) string {
	return strings.ReplaceAll(urlTemplate, URLTestServer, serverURL)
}
