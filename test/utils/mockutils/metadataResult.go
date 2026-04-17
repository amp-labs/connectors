package mockutils

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/amp-labs/connectors/common"
)

var MetadataResultComparator = metadataResultComparator{}

type metadataResultComparator struct{}

// SubsetFields checks that expected ListObjectMetadataResult fields are a subset of actual metadata result.
func (metadataResultComparator) SubsetFields(actual, expected *common.ListObjectMetadataResult) *CompareResult {
	result := NewCompareResult()
	for objectName, expectedMetadata := range expected.Result {
		actualMetadata, ok := actual.Result[objectName]
		if !ok {
			result.AddDiff(fmt.Sprintf("Result[%s] missing", objectName))
			continue
		}

		if actualMetadata.DisplayName != expectedMetadata.DisplayName {
			result.AddMismatch(fmt.Sprintf("Result[%s].DisplayName",
				objectName), expectedMetadata.DisplayName, actualMetadata.DisplayName)
		}

		for k, v := range expectedMetadata.Fields {
			value, ok := actualMetadata.Fields[k]
			if !ok {
				result.AddDiff(fmt.Sprintf("Result[%s].Fields[%s] missing", objectName, k))
				continue
			}

			if !reflect.DeepEqual(value, v) {
				result.AddMismatch(fmt.Sprintf("Result[%s].Fields[%s]",
					objectName, k), v, value)
			}
		}

		// For backwards compatibility the FieldsMap is checked alongside
		for k, v := range expectedMetadata.FieldsMap {
			value, ok := actualMetadata.FieldsMap[k]
			if !ok {
				result.AddDiff(fmt.Sprintf("Result[%s].FieldsMap[%s] missing", objectName, k))
				continue
			}

			if value != v {
				result.AddMismatch(fmt.Sprintf("Result[%s].FieldsMap[%s]",
					objectName, k), v, value)
			}
		}
	}

	return result
}

func (metadataResultComparator) SubsetErrors(actual, expected *common.ListObjectMetadataResult) *CompareResult {
	result := NewCompareResult()
	for objectName, expectedError := range expected.Errors {
		actualError, ok := actual.Errors[objectName]
		if !ok {
			result.AddDiff(fmt.Sprintf("Errors[%s] missing", objectName))
			continue
		}

		// The tester may specify ExpectedSubsetErrors with a list of errors to be present inside actualError.
		var expectedErrors ExpectedSubsetErrors
		if !errors.As(expectedError, &expectedErrors) {
			// Single expected error.
			expectedErrors = ExpectedSubsetErrors{expectedError}
		}

		if !errorsAre(actualError, expectedErrors) {
			// Subset of errors is not found under actual error for current object name.
			result.AddMismatch(fmt.Sprintf("Errors[%s]", objectName), expectedErrors, actualError)
		}
	}

	return result
}
