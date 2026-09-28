package objectcatalog

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/fileregistry"
)

// Catalog contains the canonical object metadata loaded from a static source.
//
// A Catalog is immutable after construction. Select returns a detached,
// mutable Selection that callers may enrich with dynamic metadata.
type Catalog struct {
	Objects datautils.Map[string, ObjectSpec]
}

type ObjectSpec struct {
	// Loaded from file and is served by ListObjectMetadata.
	common.ObjectMetadata
	// Object path is used to build Read URL.
	urlPath
}

type urlPath struct {
	Path string `json:"path"`
}

// MustLoad creates object metadata Catalog from static file.
func MustLoad(fileData []byte) Catalog {
	return Catalog{
		Objects: fileregistry.MustParseJSON[map[string]ObjectSpec](fileData),
	}
}

// SelectResult selects object metadata and converts the selection to the connector interface result.
func (c Catalog) SelectResult(objectNames []string) (*common.ListObjectMetadataResult, error) {
	selected, err := c.Select(objectNames)
	if err != nil {
		return nil, err
	}

	return selected.Result(), nil
}

// ReadPath returns the URL path used for reading an object.
func (c Catalog) ReadPath(objectName string) (string, error) {
	object, ok := c.Objects[objectName]
	if !ok {
		return "", common.ErrResolvingURLPathForObject
	}

	return object.Path, nil
}

// Select creates an intermediate Selection data structure which allows enriching the object metadata
// from other sources like in house custom objects, objects from provider APIs or complex cases
// where object schema is dynamic and depends on output from various provider endpoints.
func (c Catalog) Select(objectNames []string) (Selection, error) {
	if len(objectNames) == 0 {
		return Selection{}, common.ErrMissingObjects
	}

	metadata := Selection{
		objects: make(map[string]common.ObjectMetadata),
		errors:  make(map[string]error),
	}

	for _, objectName := range objectNames {
		if objectMetadata, err := c.Object(objectName); err != nil {
			metadata.errors[objectName] = err
		} else {
			metadata.objects[objectName] = *objectMetadata
		}
	}

	return metadata, nil
}

// Object returns a copy of metadata for objectName.
//
// The returned metadata can be modified without changing the catalog.
func (c Catalog) Object(objectName string) (*common.ObjectMetadata, error) {
	object, ok := c.Objects[objectName]
	if !ok {
		return nil, common.ErrObjectNotSupported
	}

	fields := copyFields(object.Fields)

	return &common.ObjectMetadata{
		DisplayName: object.DisplayName,
		Fields:      fields,
		FieldsMap:   inferDeprecatedFieldsMap(fields),
	}, nil
}

func copyFields(fields common.FieldsMetadata) common.FieldsMetadata {
	if fields == nil {
		return nil
	}

	replica := make(common.FieldsMetadata)
	for name, field := range fields {
		replica[name] = copyField(field)
	}

	return replica
}

func copyField(field common.FieldMetadata) common.FieldMetadata {
	replica := field
	if field.ReadOnly != nil {
		replica.ReadOnly = new(bool)
		*replica.ReadOnly = *field.ReadOnly
	}

	if field.IsCustom != nil {
		replica.IsCustom = new(bool)
		*replica.IsCustom = *field.IsCustom
	}

	if field.IsRequired != nil {
		replica.IsRequired = new(bool)
		*replica.IsRequired = *field.IsRequired
	}

	if field.FieldId != nil {
		replica.FieldId = new(string)
		*replica.FieldId = *field.FieldId
	}

	if field.Values != nil {
		replica.Values = make([]common.FieldValue, len(field.Values))
		copy(replica.Values, field.Values)
	}

	if field.ReferenceTo != nil {
		replica.ReferenceTo = make([]string, len(field.ReferenceTo))
		copy(replica.ReferenceTo, field.ReferenceTo)
	}

	return replica
}

func inferDeprecatedFieldsMap(fields common.FieldsMetadata) map[string]string {
	fieldsMap := make(map[string]string)

	for name, field := range fields {
		fieldsMap[name] = field.DisplayName
	}

	return fieldsMap
}
