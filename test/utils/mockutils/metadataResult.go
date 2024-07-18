package mockutils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var MetadataResultComparator = metadataResultComparator{}

type metadataResultComparator struct{}

// SubsetFields checks that expected ListObjectMetadataResult fields are a subset of actual metadata result.
func (metadataResultComparator) SubsetFields(actual, expected *common.ListObjectMetadataResult) bool {
	if len(expected.Result) == 0 {
		invalidTest("please specify expected FieldsMap response")
	}

	for objectName, expectedMetadata := range expected.Result {
		actualMetadata, ok := actual.Result[objectName]
		if !ok {
			return false
		}

		if actualMetadata.DisplayName != expectedMetadata.DisplayName {
			return false
		}

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

	mismatch := make([]error, 0)

	objectName = strings.ToLower(objectName)
	for name := range metadata.Result[objectName].FieldsMap {
		fields.CheckIfExists(name)
	}

	// every field from Read must be known to ListObjectMetadata
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
