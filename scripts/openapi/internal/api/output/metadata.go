package output

import (
	"log/slog"
	"path/filepath"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

const SchemasFile = "objectsMetadata.json"

func WriteMetadata(dirName string, pipe pipeline.Pipeline[spec.Schema]) error {
	if err := validate(pipe); err != nil {
		return err
	}

	return Write(
		filepath.Join(dirName, SchemasFile),
		extractObjectMetadata(pipe),
	)
}

func extractObjectMetadata(pipe pipeline.Pipeline[spec.Schema]) map[string]objectMetadata {
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
				DisplayName:  field.Name,
				ValueType:    getFieldValueType(field),
				ProviderType: field.Type,
				Values:       getFieldValueOptions(field),
			}
		}

		result[object.ObjectName] = objectMetadata{
			Path:        object.URLPath,
			DisplayName: object.DisplayName,
			Fields:      fields,
		}
	}

	return result
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
}

type fieldValue struct {
	Value        string `json:"value"`
	DisplayValue string `json:"displayValue"`
}

func getFieldValueType(field spec.Field) common.ValueType {
	switch field.Type {
	case "integer":
		return common.ValueTypeInt
	case "boolean":
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
