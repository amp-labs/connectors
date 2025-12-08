package deepmock

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	invopopschema "github.com/invopop/jsonschema"
	"github.com/kaptinlin/jsonschema"
)

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// injectExtrasIntoSchemaMap reads jsonschema_extras tags from a struct and injects them into the schema map.
// This is needed because the invopop library doesn't always include these custom extensions.
func injectExtrasIntoSchemaMap(schemaMap map[string]interface{}, structType reflect.Type) error {
	// Find the properties in the schema
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema has no properties to inject extras into")
	}

	// Iterate over struct fields and extract jsonschema_extras tags
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Get the JSON name for the field
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// Extract just the field name (before any options like omitempty)
		jsonName := strings.Split(jsonTag, ",")[0]

		// Get jsonschema_extras tag
		extrasTag := field.Tag.Get("jsonschema_extras")
		if extrasTag == "" {
			continue
		}

		// Parse the extras tag (format: "key1=value1,key2=value2,...")
		extras := parseExtrasTag(extrasTag)

		// Find the property in the schema and add the extras
		if propMap, ok := properties[jsonName].(map[string]interface{}); ok {
			for key, value := range extras {
				propMap[key] = value
			}
		}
	}

	return nil
}

// parseExtrasTag parses a jsonschema_extras tag value into a map of key-value pairs.
// Format: "key1=value1,key2=value2,..."
func parseExtrasTag(tag string) map[string]interface{} {
	result := make(map[string]interface{})

	// Split by comma to get individual key=value pairs
	pairs := strings.Split(tag, ",")
	for _, pair := range pairs {
		// Split by = to get key and value
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Convert "true"/"false" strings to boolean
		if value == "true" {
			result[key] = true
		} else if value == "false" {
			result[key] = false
		} else {
			result[key] = value
		}
	}

	return result
}

// filterRequiredFields filters the required array in a schema based on struct tags.
// The invopop/jsonschema library marks all fields as required by default, but we only
// want fields with the "jsonschema:required" tag to be required.
func filterRequiredFields(schemaMap map[string]interface{}, structType reflect.Type) error {
	// Build a set of fields that have the "required" tag
	requiredFields := make(map[string]bool)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Get the JSON name for the field
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// Extract just the field name (before any options like omitempty)
		jsonName := strings.Split(jsonTag, ",")[0]

		// Check for "required" in jsonschema tag
		jsonschemaTag := field.Tag.Get("jsonschema")
		if jsonschemaTag != "" {
			// Parse comma-separated tag values
			tagValues := strings.Split(jsonschemaTag, ",")
			for _, tagValue := range tagValues {
				if strings.TrimSpace(tagValue) == "required" {
					requiredFields[jsonName] = true
					break
				}
			}
		}
	}

	// Replace the required array with only explicitly required fields
	if len(requiredFields) > 0 {
		required := make([]interface{}, 0, len(requiredFields))
		for fieldName := range requiredFields {
			required = append(required, fieldName)
		}
		schemaMap["required"] = required
	} else {
		// No required fields - remove the required array entirely
		delete(schemaMap, "required")
	}

	return nil
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

		// Check if schema uses $ref and extract from $defs if needed
		var schemaMap map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return nil, fmt.Errorf("object %s: failed to unmarshal schema: %w", objectName, err)
		}

		// If schema has $ref, extract the actual schema from $defs
		if ref, hasRef := schemaMap["$ref"].(string); hasRef {
			// Parse $ref to get the definition name (format: "#/$defs/TypeName")
			parts := strings.Split(ref, "/")
			if len(parts) == 3 && parts[0] == "#" && parts[1] == "$defs" {
				defName := parts[2]
				if defs, hasDefs := schemaMap["$defs"].(map[string]interface{}); hasDefs {
					if def, hasDef := defs[defName].(map[string]interface{}); hasDef {
						// Preserve the $schema field from the root
						schemaVersion := schemaMap["$schema"]

						// Use this definition as the schema
						schemaMap = def

						// Restore the $schema field if it was present
						if schemaVersion != nil {
							schemaMap["$schema"] = schemaVersion
						}

						// Validate the extracted schema has required fields
						if _, hasType := schemaMap["type"]; !hasType {
							return nil, fmt.Errorf("object %s: extracted schema missing 'type' field", objectName)
						}
						if _, hasProps := schemaMap["properties"]; !hasProps {
							return nil, fmt.Errorf("object %s: extracted schema missing 'properties' field", objectName)
						}
					} else {
						return nil, fmt.Errorf("object %s: $ref points to non-existent definition %s", objectName, defName)
					}
				} else {
					return nil, fmt.Errorf("object %s: schema has $ref but no $defs", objectName)
				}
			} else {
				return nil, fmt.Errorf("object %s: invalid $ref format: %s", objectName, ref)
			}
		}

		// Manually inject jsonschema_extras tags into the schema map
		// This must be done after $ref extraction
		if err := injectExtrasIntoSchemaMap(schemaMap, typ); err != nil {
			return nil, fmt.Errorf("object %s: failed to inject extras: %w", objectName, err)
		}

		// Filter required fields based on struct tags
		// The invopop library marks all fields as required by default, but we only
		// want fields with the "jsonschema:required" tag to be required
		if err := filterRequiredFields(schemaMap, typ); err != nil {
			return nil, fmt.Errorf("object %s: failed to filter required fields: %w", objectName, err)
		}

		// Remove additionalProperties constraint
		// The invopop library adds "additionalProperties": false by default, which can
		// cause validation issues. We want to be lenient and allow extra properties.
		delete(schemaMap, "additionalProperties")

		// Re-marshal the processed schema
		schemaBytes, err = json.Marshal(schemaMap)
		if err != nil {
			return nil, fmt.Errorf("object %s: failed to re-marshal schema: %w", objectName, err)
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

// extractSpecialFieldsFromRaw extracts ID and updated timestamp field names from raw JSON schema.
// This is used before compilation since the compiler may not preserve custom x-amp extensions.
func extractSpecialFieldsFromRaw(rawSchema []byte) (idField, updatedField string) {
	var schemaMap map[string]any
	if err := json.Unmarshal(rawSchema, &schemaMap); err != nil {
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
		if val, exists := fieldMap["x-amp-id-field"]; exists && isTrueValue(val) {
			idField = fieldName
		}

		// Check for x-amp-updated-field extension
		if val, exists := fieldMap["x-amp-updated-field"]; exists && isTrueValue(val) {
			updatedField = fieldName
		}
	}

	return idField, updatedField
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
