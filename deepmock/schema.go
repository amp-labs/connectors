package deepmock

import (
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/kaptinlin/jsonschema"
)

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// schemaRegistry is a registry of compiled JSON schemas by object name.
type schemaRegistry map[string]*jsonschema.Schema

// Get retrieves a schema by object name.
func (r schemaRegistry) Get(objectName string) (*jsonschema.Schema, bool) {
	schema, exists := r[objectName]
	return schema, exists
}

// Set stores a schema for an object name.
func (r schemaRegistry) Set(objectName string, schema *jsonschema.Schema) {
	r[objectName] = schema
}

// parseSchemas parses raw JSON schemas into compiled schema objects.
func parseSchemas(rawSchemas map[string][]byte) (schemaRegistry, error) {
	compiler := jsonschema.NewCompiler()
	parsed := make(schemaRegistry)

	for objectName, rawSchema := range rawSchemas {
		// Compile the schema with a unique URI
		uri := fmt.Sprintf("http://deepmock.memory.store/%s", objectName)
		schema, err := compiler.Compile(rawSchema, uri)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to compile schema for %s: %w", ErrInvalidSchema, objectName, err)
		}

		parsed.Set(objectName, schema)
	}

	return parsed, nil
}

// extractSpecialFields extracts ID and updated timestamp field names from schema extensions.
func extractSpecialFields(schema *jsonschema.Schema) (idField, updatedField string) {
	if schema == nil {
		return "", ""
	}

	// Access the raw schema structure
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return "", ""
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		return "", ""
	}

	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		return "", ""
	}

	for fieldName, fieldDef := range properties {
		fieldMap, ok := fieldDef.(map[string]any)
		if !ok {
			continue
		}

		// Check for x-amp-id-field extension
		if isIDField, ok := fieldMap["x-amp-id-field"].(bool); ok && isIDField {
			idField = fieldName
		}

		// Check for x-amp-updated-field extension
		if isUpdatedField, ok := fieldMap["x-amp-updated-field"].(bool); ok && isUpdatedField {
			updatedField = fieldName
		}
	}

	return idField, updatedField
}

// validateRecord validates a record against a schema.
func validateRecord(schema *jsonschema.Schema, record map[string]any) error {
	if schema == nil {
		return fmt.Errorf("%w: schema is nil", ErrValidationFailed)
	}

	result := schema.Validate(record)
	if !result.IsValid() {
		// Collect all validation errors
		var errMessages []string
		for _, err := range result.Errors {
			errMessages = append(errMessages, err.Message)
		}
		return fmt.Errorf("%w: %v", ErrValidationFailed, errMessages)
	}

	return nil
}

// schemaToObjectMetadata converts a JSON schema to ObjectMetadata.
func schemaToObjectMetadata(objectName string, schema *jsonschema.Schema) *common.ObjectMetadata {
	if schema == nil {
		return nil
	}

	// Extract schema as map for easier processing
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		return nil
	}

	// Extract display name from title or use object name
	displayName := objectName
	if title, ok := schemaMap["title"].(string); ok && title != "" {
		displayName = title
	}

	// Extract properties
	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		properties = make(map[string]any)
	}

	// Extract required fields
	requiredFields := make(map[string]bool)
	if required, ok := schemaMap["required"].([]any); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				requiredFields[fieldName] = true
			}
		}
	}

	// Build fields metadata
	fields := make(common.FieldsMetadata)
	for fieldName, fieldDef := range properties {
		fieldMap, ok := fieldDef.(map[string]any)
		if !ok {
			continue
		}

		// Map JSON schema type to ValueType
		var valueType common.ValueType
		if typeVal, ok := fieldMap["type"].(string); ok {
			switch typeVal {
			case "string":
				valueType = common.ValueTypeString
			case "number":
				valueType = common.ValueTypeFloat
			case "integer":
				valueType = common.ValueTypeInt
			case "boolean":
				valueType = common.ValueTypeBoolean
			case "array", "object":
				// No specific type for arrays/objects, use Other
				valueType = common.ValueTypeOther
			default:
				valueType = common.ValueTypeString
			}
		}

		// Detect format-based types for date/datetime fields
		if format, ok := fieldMap["format"].(string); ok {
			switch format {
			case "date":
				valueType = common.ValueTypeDate
			case "date-time":
				valueType = common.ValueTypeDateTime
			}
		}

		// Extract enum values if present
		var values []common.FieldValue
		if enumVals, ok := fieldMap["enum"].([]any); ok && len(enumVals) > 0 {
			values = make([]common.FieldValue, 0, len(enumVals))
			for _, enumVal := range enumVals {
				// Convert enum value to string for both Value and DisplayValue
				valStr := fmt.Sprintf("%v", enumVal)
				values = append(values, common.FieldValue{
					Value:        valStr,
					DisplayValue: valStr,
				})
			}
		}

		// Map enum fields to ValueTypeSingleSelect
		if len(values) > 0 {
			valueType = common.ValueTypeSingleSelect
		}

		// Extract display name from title or use field name
		fieldDisplayName := fieldName
		if title, ok := fieldMap["title"].(string); ok && title != "" {
			fieldDisplayName = title
		}

		// Check if read-only
		var readOnly *bool
		if ro, ok := fieldMap["readOnly"].(bool); ok {
			readOnly = &ro
		}

		// Get provider type (same as JSON schema type)
		providerType := ""
		if typeVal, ok := fieldMap["type"].(string); ok {
			providerType = typeVal
		}

		// Apply required fields map
		isRequired := requiredFields[fieldName]

		fields[fieldName] = common.FieldMetadata{
			DisplayName:  fieldDisplayName,
			ValueType:    valueType,
			ProviderType: providerType,
			ReadOnly:     readOnly,
			IsRequired:   boolPtr(isRequired),
			IsCustom:     boolPtr(false),
			Values:       values,
		}
	}

	return common.NewObjectMetadata(displayName, fields)
}
