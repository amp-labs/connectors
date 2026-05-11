package mockutils

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// Any represents a wildcard value used in response comparisons.
//
// When a field value in the expected response is set to Any{},
// the comparator only checks that the field exists in the actual response,
// without validating its concrete value.
//
// Example:
//
//	expected := &common.ReadResult{
//		Data: []common.ReadResultData{
//			{
//				Raw: map[string]any{
//					"id":        Any{},
//					"createdAt": Any{},
//					"status":    "active",
//				},
//			},
//		},
//	}
//
// In this example:
//   - "id" and "createdAt" must be present in the response
//   - their values can be anything
//   - "status" must equal "active"
//
// Private field is used to make the type non-empty and prevent
// accidental structural equivalence with other empty structs.
type Any struct {
	_ any
}

var ReadResultComparator = readResultComparator{}

type readResultComparator struct{}

// SubsetRaw checks that expected ReadResult.Raw is a subset of actual ReadResult.Raw
func (readResultComparator) SubsetRaw(actual, expected *common.ReadResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff("expected at least %d data entries, got %d", len(expected.Data), len(actual.Data))
		return result
	}

	if len(expected.Data) == 0 || len(expected.Data[0].Raw) == 0 {
		invalidTest("please specify expected Raw response")
	}

	for i := range expected.Data {
		for field := range expected.Data[i].Raw {
			got, ok := actual.Data[i].Raw[field]
			exp := expected.Data[i].Raw[field]

			if _, anyValue := exp.(Any); anyValue {
				if !ok {
					result.AddDiff("Data[%d].Raw[%s] is missing", i, field)
				}
				// As long as any value is present we are good.
				continue
			}

			result.Assert(fmt.Sprintf("Data[%d].Raw[%s]", i, field), exp, got)

		}
	}

	return result
}

// SubsetFields checks that expected ReadResult.Fields is a subset of actual ReadResult.Fields
func (readResultComparator) SubsetFields(actual, expected *common.ReadResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff("expected at least %d data entries, got %d", len(expected.Data), len(actual.Data))
		return result
	}

	if len(expected.Data) == 0 || len(expected.Data[0].Fields) == 0 {
		invalidTest("please specify expected Fields response")
	}

	for i := range expected.Data {
		for field := range expected.Data[i].Fields {
			got, ok := actual.Data[i].Fields[field]
			exp := expected.Data[i].Fields[field]

			if _, anyValue := exp.(Any); anyValue {
				if !ok {
					result.AddDiff("Data[%d].Raw[%s] is missing", i, field)
				}
				// As long as any value is present we are good.
				continue
			}

			result.Assert(fmt.Sprintf("Data[%d].Fields[%s]", i, field), exp, got)
		}
	}

	return result
}

// SubsetAssociationsRaw checks that expected ReadResult.Associations are matching exactly,
// but for each Association.Raw it only checks if every mentioned expected field is present in actual raw.
func (readResultComparator) SubsetAssociationsRaw(actual, expected *common.ReadResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff("expected at least %d data entries, got %d", len(expected.Data), len(actual.Data))
		return result
	}

	for i := range expected.Data {
		message := fmt.Sprintf("Data[%d].Associations length", i)
		if !result.Assert(message, len(expected.Data[i].Associations), len(actual.Data[i].Associations)) {
			continue
		}

		for key, expectedAssociations := range expected.Data[i].Associations {
			actualAssociations, ok := actual.Data[i].Associations[key]
			if !ok {
				result.AddDiff("Data[%d].Associations[%s] missing", i, key)
				continue
			}

			message = fmt.Sprintf("Data[%d].Associations[%s] length", i, key)
			if !result.Assert(message, len(expectedAssociations), len(actualAssociations)) {
				continue
			}

			for j := range expectedAssociations {
				result.Assert(fmt.Sprintf("Data[%d].Associations[%s][%d].ObjectId", i, key, j),
					expectedAssociations[j].ObjectId, actualAssociations[j].ObjectId)

				for metaKey, exp := range expectedAssociations[j].ProviderAssociationMetadata {
					got := actualAssociations[j].ProviderAssociationMetadata[metaKey]
					message := fmt.Sprintf("Data[%d].Associations[%s][%d].ProviderAssociationMetadata[%s]", i, key, j, metaKey)
					result.Assert(message, exp, got)
				}

				// Check if expected Raw is a subset of actual Raw
				for field := range expectedAssociations[j].Raw {
					exp := expectedAssociations[j].Raw[field]
					got := actualAssociations[j].Raw[field]
					message = fmt.Sprintf("Data[%d].Associations[%s][%d].Raw[%s]", i, key, j, field)
					result.Assert(message, exp, got)
				}
			}
		}
	}

	return result
}

// Identifiers checks that actual rows have identifiers matching with expected.
// NOTE: Empty strings signify nothing should be compared.
func (c readResultComparator) Identifiers(actual *common.ReadResult,
	expected *common.ReadResult,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	for index, datum := range expected.Data {
		if index >= len(actual.Data) {
			result.AddDiff("Data[%v] does not exist, cannot check Data[%v].Id", index, index)
			break
		}

		expectedID := datum.Id
		if expectedID != "" {
			actualID := actual.Data[index].Id
			result.Assert(fmt.Sprintf("Data[%v].Id", index), expectedID, actualID)
		}
	}

	return result
}

func invalidTest(message string) {
	panic("invalid test, there is no point to check if empty set belongs to any set; " + message)
}
