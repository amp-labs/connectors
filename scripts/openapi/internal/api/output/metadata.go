package output

import (
	"encoding/json"
	"log/slog"
	"maps"
	"path/filepath"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

const (
	ObjectMetadataFile    = "objectsMetadata.json"
	ObjectMetadataZipFile = "objectsMetadata.json.gz"
)

func WriteMetadata(dirName string, pipe pipeline.Pipeline[spec.Schema], options ...MetadataOption) error {
	return writeMetadata(dirName, ObjectMetadataFile, pipe, options...)
}

func WriteZippedMetadata(dirName string, pipe pipeline.Pipeline[spec.Schema], options ...MetadataOption) error {
	return writeMetadata(dirName, ObjectMetadataZipFile, pipe, options...)
}

func writeMetadata(dirName, fileName string, pipe pipeline.Pipeline[spec.Schema], options ...MetadataOption) error {
	if err := validate(pipe); err != nil {
		return err
	}

	params := createMetadataParams(options)

	metadata, err := extractObjectMetadata(params, pipe)
	if err != nil {
		return err
	}

	return Write(
		filepath.Join(dirName, fileName),
		metadata,
	)
}

func extractObjectMetadata(params metadataParams,
	pipe pipeline.Pipeline[spec.Schema],
) (map[string]objectMetadata, error) {
	result := make(map[string]objectMetadata)

	for _, object := range pipe.List() {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		fields := make(map[string]fieldMetadata)
		for _, field := range object.Fields {
			fields[field.Name] = fieldMetadata{
				DisplayName:  field.DisplayName,
				ValueType:    getFieldValueType(field),
				ProviderType: field.Type,
				Values:       getFieldValueOptions(field),
				ReferenceTo:  field.ReferenceTo,
			}
		}

		result[object.ObjectName] = objectMetadata{
			Path:        object.UrlPath,
			DisplayName: object.DisplayName,
			Fields:      fields,
		}
	}

	if len(params.manualObjects) > 0 {
		manualObjects, err := parseManualObjects(params.manualObjects)
		if err != nil {
			return nil, err
		}

		maps.Copy(result, manualObjects)
	}

	return result, nil
}

func parseManualObjects(data []byte) (map[string]objectMetadata, error) {
	var result map[string]objectMetadata
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// This matches common.ObjectMetadata.
type objectMetadata struct {
	// extra field not present in common.ObjectMetadata.
	Path        string                   `json:"path"`
	DisplayName string                   `json:"displayName"`
	Fields      map[string]fieldMetadata `json:"fields"`
}

type fieldMetadata struct {
	DisplayName  string           `json:"displayName"`
	ValueType    common.ValueType `json:"valueType"`
	ProviderType string           `json:"providerType,omitempty"`
	Values       []fieldValue     `json:"values,omitempty"`
	ReferenceTo  []string         `json:"referenceTo,omitempty"`
}

type fieldValue struct {
	Value        string `json:"value"`
	DisplayValue string `json:"displayValue"`
}

func getFieldValueType(field spec.Field) common.ValueType {
	if field.ValueType != "" {
		return field.ValueType
	}

	if len(field.ReferenceTo) != 0 {
		return common.ValueTypeReference
	}

	switch field.Type {
	case "integer":
		return common.ValueTypeInt
	case "boolean", "bool":
		return common.ValueTypeBoolean
	case "string":
		if len(field.ValueOptions) != 0 {
			return common.ValueTypeSingleSelect
		}

		return common.ValueTypeString
	default:
		// object, array
		return common.ValueTypeOther
	}
}

func getFieldValueOptions(field spec.Field) []fieldValue {
	if len(field.ValueOptions) == 0 {
		return nil
	}

	values := make([]fieldValue, len(field.ValueOptions))
	for index, option := range field.ValueOptions {
		values[index] = fieldValue{
			Value:        option,
			DisplayValue: option,
		}
	}

	return values
}

type metadataParams struct {
	manualObjects []byte
}

// MetadataOption configures how object metadata is written to a file.
type MetadataOption func(params *metadataParams)

func createMetadataParams(opts []MetadataOption) metadataParams {
	var params metadataParams
	for _, opt := range opts {
		opt(&params)
	}

	return params
}

// WithManualObjects adds manually defined object metadata to the generated
// metadata before writing it to a file.
//
// The provided data must be a JSON object mapping object names to their
// metadata. Manual entries take precedence over generated entries with the
// same object name.
//
// Returns an error if the provided JSON cannot be parsed.
func WithManualObjects(manualObjects []byte) MetadataOption {
	return func(params *metadataParams) {
		params.manualObjects = manualObjects
	}
}
