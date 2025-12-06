package deepmock

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/amp-labs/connectors/common"
	invopopschema "github.com/invopop/jsonschema"
	"github.com/kaptinlin/jsonschema"
)

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// DeriveSchemasFromStructs converts Go structs to JSON Schema bytes for use with NewConnector.
//
// This function provides a developer-friendly API for defining schemas using Go types with
// struct tags, as an alternative to writing raw JSON schemas. The generated schemas are
// compatible with the Draft 2020-12 specification and can be used directly with NewConnector.
//
// Parameters:
//   - schemas: Map of object names to struct instances (e.g., &User{}, &Contact{})
//
// Returns:
//   - Map of object names to JSON Schema bytes (Draft 2020-12 format)
//   - Error if any struct is invalid or schema generation fails
//
// Custom Extensions:
//
// Use the jsonschema_extras struct tag to add custom x-amp-* extensions:
//   - x-amp-id-field: Marks a field as the unique identifier
//   - x-amp-updated-field: Marks a field as the last updated timestamp
//
// Standard jsonschema tags are also supported:
//   - required: Mark field as required
//   - enum: Define allowed values
//   - format: Specify format (email, date, date-time, etc.)
//   - minLength, maxLength: String length constraints
//   - minimum, maximum: Numeric constraints
//   - title: Display name for field
//   - description: Field description
//
// Example:
//
//	type Contact struct {
//	    ID        string    `json:"id" jsonschema_extras:"x-amp-id-field=true"`
//	    UpdatedAt time.Time `json:"updated_at" jsonschema_extras:"x-amp-updated-field=true"`
//	    Name      string    `json:"name" jsonschema:"required,title=Full Name"`
//	    Email     string    `json:"email" jsonschema:"required,format=email"`
//	    Status    string    `json:"status" jsonschema:"enum=active,enum=inactive"`
//	}
//
//	type Company struct {
//	    ID        int64  `json:"id" jsonschema_extras:"x-amp-id-field=true"`
//	    UpdatedAt int64  `json:"updated_at" jsonschema_extras:"x-amp-updated-field=true"`
//	    Name      string `json:"name" jsonschema:"required,minLength=1,maxLength=100"`
//	}
//
//	schemas, err := DeriveSchemasFromStructs(map[string]interface{}{
//	    "contacts":  &Contact{},
//	    "companies": &Company{},
//	})
//	if err != nil {
//	    // handle error
//	}
//
//	connector, err := NewConnector(schemas)
//	if err != nil {
//	    // handle error
//	}
//
// Error Handling:
//
// The function returns an error if:
//   - Input map is nil or empty
//   - Any value is not a struct or pointer to struct
//   - Schema generation fails for any struct
//   - JSON marshaling fails
func DeriveSchemasFromStructs(schemas map[string]interface{}) (map[string][]byte, error) {
	// Validate input
	if schemas == nil || len(schemas) == 0 {
		return nil, fmt.Errorf("schemas map cannot be nil or empty")
	}

	result := make(map[string][]byte)

	for objectName, structInstance := range schemas {
		// Validate that the value is a struct or pointer to struct
		val := reflect.ValueOf(structInstance)

		// Check if the value is valid (i.e., not nil)
		if !val.IsValid() {
			return nil, fmt.Errorf("object %s: struct instance cannot be nil", objectName)
		}

		// Check for typed nil pointers (e.g., (*MyStruct)(nil))
		if val.Kind() == reflect.Ptr && val.IsNil() {
			return nil, fmt.Errorf("object %s: struct instance cannot be nil", objectName)
		}

		typ := val.Type()

		// Dereference pointer if needed
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		// Validate it's a struct
		if typ.Kind() != reflect.Struct {
			return nil, fmt.Errorf("object %s: expected struct or pointer to struct, got %s", objectName, typ.Kind())
		}

		// Create a new Reflector instance for this struct
		reflector := new(invopopschema.Reflector)

		// Generate the schema
		schema := reflector.Reflect(structInstance)
		if schema == nil {
			return nil, fmt.Errorf("object %s: failed to generate schema (reflector returned nil)", objectName)
		}

		// Marshal schema to JSON bytes
		schemaBytes, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("object %s: failed to marshal schema to JSON: %w", objectName, err)
		}

		// Store in result map
		result[objectName] = schemaBytes
	}

	return result, nil
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

// isTrueValue checks if a value represents a "true" boolean or string.
// Returns true for:
//   - boolean true
//   - string "true" (case-insensitive, trimmed)
func isTrueValue(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		// Trim spaces and check case-insensitive
		trimmed := ""
		for _, r := range v {
			if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
				trimmed += string(r)
			}
		}
		return trimmed == "true" || trimmed == "TRUE" || trimmed == "True"
	default:
		return false
	}
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

		// Check for x-amp-id-field extension (supports both boolean and string "true")
		if val, exists := fieldMap["x-amp-id-field"]; exists && isTrueValue(val) {
			idField = fieldName
		}

		// Check for x-amp-updated-field extension (supports both boolean and string "true")
		if val, exists := fieldMap["x-amp-updated-field"]; exists && isTrueValue(val) {
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
