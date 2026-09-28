package objectcatalog

import (
	"github.com/amp-labs/connectors/common"
)

// Selection is a request-scoped, mutable collection of object metadata.
//
// It is initialized from a Catalog and may be enriched with metadata from
// custom objects or other dynamic sources.
type Selection struct {
	objects map[string]common.ObjectMetadata
	errors  map[string]error
}

func (s Selection) AddField(objectName, fieldName string, metadata common.FieldMetadata) bool {
	object, ok := s.objects[objectName]
	if !ok {
		return false
	}

	object.Fields[fieldName] = metadata
	object.FieldsMap[fieldName] = metadata.DisplayName

	return true
}

// Result converts the selection to the output expected by connectors.ObjectMetadataConnector.
func (s Selection) Result() *common.ListObjectMetadataResult {
	return &common.ListObjectMetadataResult{
		Result: s.objects,
		Errors: s.errors,
	}
}
