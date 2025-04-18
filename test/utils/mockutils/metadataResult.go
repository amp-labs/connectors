package mockutils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

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

		// For backwards compatability the FieldsMap is checked alongside
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

// ValidateReadConformsMetadata this will check that all the fields that were returned by `Read` method
// are a subset of ObjectMetadata. It is possible that Read will not return all the possible fields
// which is fine and not a cause for an error.
// However, it will return a Joined error for every Read field that is missing in Metadata.
func ValidateReadConformsMetadata(objectName string,
	readResponse map[string]any, metadata *common.ListObjectMetadataResult,
) error {
	fields := make(checkList)

	for field := range readResponse {
		fields.Add(field)
	}

	for name := range metadata.Result[objectName].FieldsMap {
		fields.CheckIfExists(name)
	}

	// every field from Read must be known to ListObjectMetadata
	mismatch := make([]error, 0)

	for name, checked := range fields {
		if !checked {
			mismatch = append(mismatch, fmt.Errorf("metadata schema is missing field %v", name))
		}
	}

	return errors.Join(mismatch...)
}

type checkList map[string]bool

func (l checkList) Add(value string) {
	l[strings.ToLower(value)] = false
}

func (l checkList) Has(value string) bool {
	_, found := l[strings.ToLower(value)]

	return found
}

func (l checkList) CheckIfExists(value string) {
	if l.Has(value) {
		l[strings.ToLower(value)] = true
	}
}
