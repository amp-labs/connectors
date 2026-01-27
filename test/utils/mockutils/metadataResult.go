package mockutils

import (
	"errors"
	"reflect"

	"github.com/amp-labs/connectors/common"
)

var MetadataResultComparator = metadataResultComparator{}

type metadataResultComparator struct{}

// SubsetFields checks that expected ListObjectMetadataResult fields are a subset of actual metadata result.
func (metadataResultComparator) SubsetFields(actual, expected *common.ListObjectMetadataResult) bool {
	for objectName, expectedMetadata := range expected.Result {
		actualMetadata, ok := actual.Result[objectName]
		if !ok {
			return false
		}

		if actualMetadata.DisplayName != expectedMetadata.DisplayName {
			return false
		}

		for k, v := range expectedMetadata.Fields {
			value, ok := actualMetadata.Fields[k]
			if !ok {
				return false
			}

			if !reflect.DeepEqual(value, v) {
				return false
			}
		}

		// For backwards compatibility the FieldsMap is checked alongside
		for k, v := range expectedMetadata.FieldsMap {
			value, ok := actualMetadata.FieldsMap[k]
			if !ok {
				return false
			}

			if value != v {
				return false
			}
		}
	}

	return true
}

func (metadataResultComparator) SubsetErrors(actual, expected *common.ListObjectMetadataResult) bool {
	for objectName, expectedError := range expected.Errors {
		actualError, ok := actual.Errors[objectName]
		if !ok {
			return false
		}

		// The tester may specify ExpectedSubsetErrors with a list of errors to be present inside actualError.
		var expectedErrors ExpectedSubsetErrors
		if !errors.As(expectedError, &expectedErrors) {
			// Single expected error.
			expectedErrors = ExpectedSubsetErrors{expectedError}
		}

		if !errorsAre(actualError, expectedErrors) {
			// Subset of errors is not found under actual error for current object name.
			// No need to check other object names.
			return false
		}
	}

	return true
}
