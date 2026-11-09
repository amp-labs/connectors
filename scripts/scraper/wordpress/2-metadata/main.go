package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/scraper/wordpress/internal/files"
)

type InputFile map[string]DiscoveryResponse

// DiscoveryResponse is returned when using OPTIONS HTTP method on any endpoint.
type DiscoveryResponse struct {
	Schema ObjectSchema `json:"schema"`
}

type ObjectSchema struct {
	Properties map[string]ObjectProperty `json:"properties"`
}

type ObjectProperty struct {
	// Type can be an array of strings or a string.
	Type any `json:"type,omitempty"`
	// The type can be a string, but format is date-time, then
	// we can treat this property as a date.
	Format string `json:"format,omitempty"`
	// Enum list of options for a type.
	Enum []any `json:"enum,omitempty"`
}

type OutputType struct {
	DisplayName string                 `json:"displayName"`
	Path        string                 `json:"path"`
	Fields      map[string]OutputField `json:"fields"`
}

type OutputField struct {
	DisplayName  string           `json:"displayName"`
	ProviderType string           `json:"providerType"`
	ValueType    common.ValueType `json:"valueType"`
	Values       []FieldValue     `json:"values,omitempty"`
}

type FieldValue struct {
	Value        string `json:"value"`
	DisplayValue string `json:"displayValue"`
}

func main() {
	var input InputFile
	goutils.MustBeNil(json.Unmarshal(files.CombinedResponses, &input))

	result := make(map[string]OutputType)

	for path, item := range input {
		parts := strings.Split(path, "/")
		objectName := parts[len(parts)-1]

		// Some exceptions to compound naming to avoid collisions.
		switch path {
		case "/wp/v2/sites/{{workspace_id}}/block-patterns/patterns":
			objectName = "block-patterns/patterns"
		case "/wp/v2/sites/{{workspace_id}}/pattern-directory/patterns":
			objectName = "pattern-directory/patterns"
		}

		fields := make(map[string]OutputField)
		for propertyName, property := range item.Schema.Properties {
			fieldValues := make([]FieldValue, len(property.Enum))
			for idx, enum := range property.Enum {
				fieldValues[idx] = FieldValue{
					Value:        fmt.Sprintf("%v", enum),
					DisplayValue: fmt.Sprintf("%v", enum),
				}
			}

			var propertyType string

			switch val := property.Type.(type) {
			case []any:
				array := convertToSliceString(val)
				sort.Strings(array)
				propertyType = strings.Join(array, "|")
			case string:
				propertyType = val
			default:
				fmt.Println("property type is neither a string nor a slice") // nolint:forbidigo
			}

			if property.Format != "" {
				propertyType = propertyType + "(" + property.Format + ")"
			}

			fields[propertyName] = OutputField{
				DisplayName:  displayNameFormat(propertyName),
				ProviderType: propertyType,
				ValueType:    providerTypeToValueType(propertyType),
				Values:       fieldValues,
			}
		}

		result[objectName] = OutputType{
			DisplayName: displayNameFormat(objectName),
			Path:        path,
			Fields:      fields,
		}
	}

	goutils.MustBeNil(codec.WriteJsonZip(files.WordPressMetadataFile, result))
}

func convertToSliceString(val []any) []string {
	result := make([]string, len(val))
	for i, v := range val {
		var ok bool

		result[i], ok = v.(string)
		if !ok {
			fmt.Println("slice item is not a string") // nolint:forbidigo
		}
	}

	return result
}

func providerTypeToValueType(providerType string) common.ValueType {
	switch providerType {
	case "bool", "boolean", "boolean|string":
		return common.ValueTypeBoolean
	case "integer", "integer|null", "number":
		return common.ValueTypeInt
	case "string(date-time)", "null|string(date-time)":
		return common.ValueTypeDateTime
	case "string", "string(ip)", "string(uri)", "string(email)", "null|string":
		return common.ValueTypeString
	case "object", "object|string", "null|object", "array", "array|null":
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

func displayNameFormat(name string) string {
	result := name
	for _, transform := range []func(string) string{
		naming.SeparateUnderscoreWords,
		naming.SeparateKebabWords,
		naming.SeparateCamelCaseWords,
		naming.CapitalizeFirstLetterEveryWord,
	} {
		result = transform(result)
	}

	return result
}
