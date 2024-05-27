package mockutils

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// ValidateReadConformsMetadata this will check that all the fields that were returned by `Read` method
// are a subset of ObjectMetadata. It is possible that Read will not return all the possible fields
// which is fine and not a cause for an error.
// However, it will return a Joined error for every Read field that is missing in Metadata.
func ValidateReadConformsMetadata(objectName string,
	readResponse map[string]any, metadata *common.ListObjectMetadataResult) error {
	fields := make(map[string]bool)

	for field := range readResponse {
		fields[field] = false
	}

	mismatch := make([]error, 0)

	for _, displayName := range metadata.Result[objectName].FieldsMap {
		if _, found := fields[displayName]; found {
			fields[displayName] = true
		}
	}

	// every field from Read must be known to ListObjectMetadata
	for name, found := range fields {
		if !found {
			mismatch = append(mismatch, fmt.Errorf("metadata schema is missing field %v", name))
		}
	}

	return errors.Join(mismatch...)
}
