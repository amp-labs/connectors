package mockutils

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var ReadResultComparator = readResultComparator{}

type readResultComparator struct{}

// SubsetRaw checks that expected ReadResult.Raw is a subset of actual ReadResult.Raw
func (readResultComparator) SubsetRaw(actual, expected *common.ReadResult) *CompareResult {
	result := NewCompareResult()
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff(fmt.Sprintf("expected at least %d data entries, got %d", len(expected.Data), len(actual.Data)))
		return result
	}

	if len(expected.Data) == 0 || len(expected.Data[0].Raw) == 0 {
		invalidTest("please specify expected Raw response")
	}

	for i := range expected.Data {
		for field := range expected.Data[i].Raw {
			got := actual.Data[i].Raw[field]
			exp := expected.Data[i].Raw[field]
			result.AddMismatch(fmt.Sprintf("Data[%d].Raw[%s]", i, field), got, exp)
		}
	}

	return result
}

// SubsetFields checks that expected ReadResult.Fields is a subset of actual ReadResult.Fields
func (readResultComparator) SubsetFields(actual, expected *common.ReadResult) *CompareResult {
	result := NewCompareResult()
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff(fmt.Sprintf("expected at least %d data entries, got %d", len(expected.Data), len(actual.Data)))
		return result
	}

	if len(expected.Data) == 0 || len(expected.Data[0].Fields) == 0 {
		invalidTest("please specify expected Fields response")
	}

	for i := range expected.Data {
		for field := range expected.Data[i].Fields {
			got := actual.Data[i].Fields[field]
			exp := expected.Data[i].Fields[field]
			result.AddMismatch(fmt.Sprintf("Data[%d].Fields[%s]", i, field), got, exp)
		}
	}

	return result
}

// SubsetAssociationsRaw checks that expected ReadResult.Associations are matching exactly,
// but for each Association.Raw it only checks if every mentioned expected field is present in actual raw.
func (readResultComparator) SubsetAssociationsRaw(actual, expected *common.ReadResult) *CompareResult {
	result := NewCompareResult()
	if len(actual.Data) < len(expected.Data) {
		result.AddDiff(fmt.Sprintf("expected at least %d data entries, got %d", len(expected.Data), len(actual.Data)))
		return result
	}

	for i := range expected.Data {
		message := fmt.Sprintf("Data[%d].Associations length", i)
		if result.AddMismatch(message, len(expected.Data[i].Associations), len(actual.Data[i].Associations)) {
			continue
		}

		for key, expectedAssociations := range expected.Data[i].Associations {
			actualAssociations, ok := actual.Data[i].Associations[key]
			if !ok {
				result.AddDiff(fmt.Sprintf("Data[%d].Associations[%s] missing", i, key))
				continue
			}

			message = fmt.Sprintf("Data[%d].Associations[%s] length", i, key)
			if result.AddMismatch(message, len(expectedAssociations), len(actualAssociations)) {
				continue
			}

			for j := range expectedAssociations {
				message = fmt.Sprintf("Data[%d].Associations[%s][%d].ObjectId", i, key, j)
				result.AddMismatch(
					message, expectedAssociations[j].ObjectId, actualAssociations[j].ObjectId,
				)

				message = fmt.Sprintf("Data[%d].Associations[%s][%d].AssociationType", i, key, j)
				result.AddMismatch(
					message, expectedAssociations[j].AssociationType, actualAssociations[j].AssociationType,
				)

				// Check if expected Raw is a subset of actual Raw
				for field := range expectedAssociations[j].Raw {
					exp := expectedAssociations[j].Raw[field]
					got := actualAssociations[j].Raw[field]
					message = fmt.Sprintf("Data[%d].Associations[%s][%d].Raw[%s]", i, key, j, field)
					result.AddMismatch(message, got, exp)
				}
			}
		}
	}

	return result
}

// Identifiers checks that actual rows have identifiers matching with expected.
// Empty strings signify nothing should be compared.
func (c readResultComparator) Identifiers(actual *common.ReadResult, expected *common.ReadResult) *CompareResult {
	result := NewCompareResult()
	for index, datum := range expected.Data {
		expectedID := datum.Id
		if expectedID != "" {
			actualID := actual.Data[index].Id
			if actualID != expectedID {
				result.AddMismatch(fmt.Sprintf("Data[%v].Id", index), actualID, expectedID)
			}
		}
	}

	return result
}

func invalidTest(message string) {
	panic("invalid test, there is no point to check if empty set belongs to any set; " + message)
}
