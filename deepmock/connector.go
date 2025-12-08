package deepmock

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// Connector is an in-memory mock connector with JSON schema validation.
type Connector struct {
	client  *common.JSONHTTPClient
	params  *parameters
	schemas schemaRegistry
	storage *Storage
}

// Compile-time interface checks.
var (
	_ connectors.Connector               = (*Connector)(nil)
	_ connectors.ReadConnector           = (*Connector)(nil)
	_ connectors.WriteConnector          = (*Connector)(nil)
	_ connectors.DeleteConnector         = (*Connector)(nil)
	_ connectors.ObjectMetadataConnector = (*Connector)(nil)
)

// NewConnector creates a new deepmock connector instance.
func NewConnector(schemas map[string][]byte, opts ...Option) (*Connector, error) {
	// Apply options without pre-populated schemas/storage
	params, err := paramsbuilder.Apply(parameters{}, opts, WithClient(http.DefaultClient))
	if err != nil {
		return nil, err
	}

	// Determine which schemas to use (raw vs struct-derived)
	var finalSchemas map[string][]byte

	if schemas != nil && len(schemas) > 0 {
		// Raw schemas provided
		finalSchemas = schemas

		// Warn if both raw and struct schemas are provided
		if params.structSchemas != nil && len(params.structSchemas) > 0 {
			slog.Warn("both raw schemas and struct schemas provided; using raw schemas",
				"rawSchemaCount", len(schemas),
				"structSchemaCount", len(params.structSchemas))
		}
	} else if params.structSchemas != nil && len(params.structSchemas) > 0 {
		// Derive schemas from structs
		var err error
		finalSchemas, err = DeriveSchemasFromStructs(params.structSchemas)
		if err != nil {
			return nil, fmt.Errorf("failed to derive schemas from structs: %w", err)
		}
	} else {
		// Neither provided
		return nil, fmt.Errorf("%w: must provide either raw schemas or use WithStructSchemas option", ErrMissingParam)
	}

	// Extract special fields from raw schemas before compilation
	// (compilation may not preserve custom x-amp extensions)
	idFields := make(map[string]string)
	updatedFields := make(map[string]string)
	for objectName, rawSchema := range finalSchemas {
		idField, updatedField := extractSpecialFieldsFromRaw(rawSchema)
		if idField != "" {
			idFields[objectName] = idField
		}
		if updatedField != "" {
			updatedFields[objectName] = updatedField
		}
	}

	// Parse schemas
	parsedSchemas, err := parseSchemas(finalSchemas)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schemas: %w", err)
	}

	// Initialize storage with parsed schemas and special fields
	storage := NewStorage(parsedSchemas, idFields, updatedFields)

	return &Connector{
		client: &common.JSONHTTPClient{
			HTTPClient: params.Caller,
		},
		params:  params,
		schemas: parsedSchemas,
		storage: storage,
	}, nil
}

// String returns the connector name.
func (c *Connector) String() string {
	return "deepmock"
}

// JSONHTTPClient returns the JSON HTTP client.
func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.client
}

// HTTPClient returns the HTTP client.
func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.client.HTTPClient
}

// Provider returns the provider information.
func (c *Connector) Provider() providers.Provider {
	return providers.DeepMock
}

// Read retrieves records for an object with pagination and filtering.
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	// Validate parameters
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	// Check if object schema exists
	if _, exists := c.schemas.Get(params.ObjectName); !exists {
		return nil, fmt.Errorf("%w: %s", ErrSchemaNotFound, params.ObjectName)
	}

	// Get records from storage with time filtering
	records, err := c.storage.List(params.ObjectName, params.Since, params.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	// Apply field filtering if specified
	if len(params.Fields) > 0 {
		filteredRecords := make([]map[string]any, len(records))
		for i, record := range records {
			filtered := make(map[string]any)
			for field := range params.Fields {
				if value, exists := record[field]; exists {
					filtered[field] = value
				}
			}
			filteredRecords[i] = filtered
		}
		records = filteredRecords
	}

	// Parse pagination parameters
	offset := 0
	if params.NextPage != "" {
		var err error
		offset, err = strconv.Atoi(string(params.NextPage))
		if err != nil {
			offset = 0
		}
	}

	pageSize := 100
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	// Apply pagination
	start := offset
	end := offset + pageSize
	if start > len(records) {
		start = len(records)
	}
	if end > len(records) {
		end = len(records)
	}

	pageRecords := records[start:end]

	// Build result rows
	rows := make([]common.ReadResultRow, len(pageRecords))
	for i, record := range pageRecords {
		// Build fields map with lowercase keys
		fields := make(map[string]any)
		for key, value := range record {
			fields[strings.ToLower(key)] = value
		}

		rows[i] = common.ReadResultRow{
			Fields: fields,
			Raw:    record,
		}
	}

	// Calculate next page token
	var nextPage common.NextPageToken
	done := true
	if end < len(records) {
		nextPage = common.NextPageToken(strconv.Itoa(end))
		done = false
	}

	return &common.ReadResult{
		Rows:     int64(len(pageRecords)),
		Data:     rows,
		NextPage: nextPage,
		Done:     done,
	}, nil
}

// Write creates or updates a record.
func (c *Connector) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	// Validate parameters
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	// Check if object schema exists
	schema, exists := c.schemas.Get(params.ObjectName)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSchemaNotFound, params.ObjectName)
	}

	// Convert record data to map
	recordMap, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert record data: %w", err)
	}

	var recordID string
	var finalRecord map[string]any

	// Determine operation (create vs update)
	if params.RecordId == "" {
		// CREATE operation
		idField := c.storage.idFields[params.ObjectName]
		updatedField := c.storage.updatedFields[params.ObjectName]

		// Generate ID if field exists and not provided
		if idField != "" {
			if _, exists := recordMap[idField]; !exists {
				recordMap[idField] = generateID(schema, idField)
			}
			// Extract record ID
			recordID = fmt.Sprintf("%v", recordMap[idField])
		} else {
			// No ID field, generate UUID
			recordID = uuid.New().String()
		}

		// Generate timestamp for updated field if not provided
		if updatedField != "" {
			if _, exists := recordMap[updatedField]; !exists {
				recordMap[updatedField] = generateTimestamp(schema, updatedField)
			}
		}

		finalRecord = recordMap
	} else {
		// UPDATE operation
		recordID = params.RecordId

		// Retrieve existing record
		existing, err := c.storage.Get(params.ObjectName, recordID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve existing record: %w", err)
		}

		// Merge new data with existing
		for key, value := range recordMap {
			existing[key] = value
		}

		// Update timestamp field if not explicitly provided in the update
		updatedField := c.storage.updatedFields[params.ObjectName]
		if updatedField != "" {
			// Only auto-generate if not explicitly provided in the update data
			if _, providedInUpdate := recordMap[updatedField]; !providedInUpdate {
				existing[updatedField] = generateTimestamp(schema, updatedField)
			}
		}

		finalRecord = existing
	}

	// Validate record against schema
	if err := validateRecord(schema, finalRecord); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Store record
	if err := c.storage.Store(params.ObjectName, recordID, finalRecord); err != nil {
		return nil, fmt.Errorf("failed to store record: %w", err)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     finalRecord,
	}, nil
}

// Delete removes a record.
func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	// Validate parameters
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	// Check if object schema exists
	if _, exists := c.schemas.Get(params.ObjectName); !exists {
		return nil, fmt.Errorf("%w: %s", ErrSchemaNotFound, params.ObjectName)
	}

	// Delete record from storage
	err := c.storage.Delete(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	return &connectors.DeleteResult{
		Success: true,
	}, nil
}

// ListObjectMetadata returns metadata for specified objects.
func (c *Connector) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, fmt.Errorf("%w: objectNames", ErrMissingParam)
	}

	result := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		schema, exists := c.schemas.Get(objectName)
		if !exists {
			result.AppendError(objectName, fmt.Errorf("%w: %s", ErrSchemaNotFound, objectName))
			continue
		}

		metadata := schemaToObjectMetadata(objectName, schema)
		if metadata == nil {
			result.AppendError(objectName, fmt.Errorf("failed to convert schema to metadata for %s", objectName))
			continue
		}

		result.Result[objectName] = *metadata
	}

	return result, nil
}

// GenerateRandomRecord generates a random record conforming to the object's schema.
func (c *Connector) GenerateRandomRecord(objectName string) (map[string]any, error) {
	return c.generateRandomRecordWithDepth(objectName, 0)
}

// generateRandomRecordWithDepth generates a random record with depth limiting for recursion.
func (c *Connector) generateRandomRecordWithDepth(objectName string, depth int) (map[string]any, error) {
	const maxDepth = 5
	const maxRetries = 100

	schema, exists := c.schemas.Get(objectName)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSchemaNotFound, objectName)
	}

	// Extract schema as map
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		properties = make(map[string]any)
	}

	// Get special fields
	idField := c.storage.idFields[objectName]
	updatedField := c.storage.updatedFields[objectName]

	// Retry logic for validation failures
	var record map[string]any
	var validationErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		record = make(map[string]any)

		// Generate values for each property
		for fieldName, fieldDef := range properties {
			fieldMap, ok := fieldDef.(map[string]any)
			if !ok {
				continue
			}

			// Handle special fields
			if fieldName == idField {
				record[fieldName] = generateID(schema, idField)
				continue
			}
			if fieldName == updatedField {
				record[fieldName] = generateTimestamp(schema, updatedField)
				continue
			}

			// Generate value based on field schema
			value, err := c.generateFieldValue(fieldMap, depth, maxDepth)
			if err != nil {
				// Log error with field name and schema details
				fieldSchemaJSON, _ := json.Marshal(fieldMap)
				slog.Error("field value generation failed, using fallback",
					"objectName", objectName,
					"fieldName", fieldName,
					"error", err,
					"fieldSchema", string(fieldSchemaJSON),
					"attempt", attempt+1,
				)
				record[fieldName] = gofakeit.Word()
				continue
			}
			record[fieldName] = value
		}

		// Validate generated record
		validationErr = validateRecord(schema, record)
		if validationErr == nil {
			return record, nil
		}
	}

	// All retries failed
	return nil, fmt.Errorf("generated record validation failed after %d attempts: %w", maxRetries, validationErr)
}

// generateFieldValue generates a value for a field based on its schema definition.
func (c *Connector) generateFieldValue(fieldMap map[string]any, depth, maxDepth int) (any, error) {
	fieldType, _ := fieldMap["type"].(string)
	format, _ := fieldMap["format"].(string)

	// Check for enum
	if enumValues, ok := fieldMap["enum"].([]any); ok && len(enumValues) > 0 {
		return enumValues[gofakeit.Number(0, len(enumValues)-1)], nil
	}

	switch fieldType {
	case "string":
		return c.generateStringValue(fieldMap, format)
	case "integer":
		return c.generateIntegerValue(fieldMap)
	case "number":
		return c.generateNumberValue(fieldMap)
	case "boolean":
		return gofakeit.Bool(), nil
	case "array":
		return c.generateArrayValue(fieldMap, depth, maxDepth)
	case "object":
		return c.generateObjectValue(fieldMap, depth, maxDepth)
	default:
		return gofakeit.Word(), nil
	}
}

// generateStringValue generates a string value with pattern and length constraints.
func (c *Connector) generateStringValue(fieldMap map[string]any, format string) (string, error) {
	const maxAttempts = 3

	// Check for pattern
	if pattern, ok := fieldMap["pattern"].(string); ok && pattern != "" {
		// Use regex pattern to generate matching string
		for attempt := 0; attempt < maxAttempts; attempt++ {
			value := gofakeit.Regex(pattern)
			if c.validateStringConstraints(value, fieldMap) {
				return value, nil
			}
		}
		// If pattern generation fails, fall through to format/length-based generation
	}

	// Extract length constraints
	var minLength, maxLength *int
	if ml, ok := fieldMap["minLength"].(float64); ok {
		minLen := int(ml)
		minLength = &minLen
	}
	if ml, ok := fieldMap["maxLength"].(float64); ok {
		maxLen := int(ml)
		maxLength = &maxLen
	}

	// Generate by format
	for attempt := 0; attempt < maxAttempts; attempt++ {
		value := generateStringByFormat(format, minLength, maxLength)
		if c.validateStringConstraints(value, fieldMap) {
			return value, nil
		}
	}

	// Fallback: generate string with length constraints
	return c.generateStringWithLength(minLength, maxLength), nil
}

// validateStringConstraints checks if a string meets length constraints.
func (c *Connector) validateStringConstraints(value string, fieldMap map[string]any) bool {
	if minLength, ok := fieldMap["minLength"].(float64); ok {
		if len(value) < int(minLength) {
			return false
		}
	}
	if maxLength, ok := fieldMap["maxLength"].(float64); ok {
		if len(value) > int(maxLength) {
			return false
		}
	}
	return true
}

// generateStringWithLength generates a string meeting length constraints.
func (c *Connector) generateStringWithLength(minLength, maxLength *int) string {
	min := 1
	max := 50
	if minLength != nil && *minLength > 0 {
		min = *minLength
	}
	if maxLength != nil && *maxLength > 0 {
		max = *maxLength
	}
	if max < min {
		max = min
	}

	// Generate target length
	targetLength := gofakeit.Number(min, max)

	// Generate string of appropriate length
	if targetLength <= 20 {
		return gofakeit.LetterN(uint(targetLength))
	}

	// For longer strings, use sentence and trim/pad as needed
	sentence := gofakeit.Sentence(targetLength / 5)
	if len(sentence) > targetLength {
		return sentence[:targetLength]
	}
	if len(sentence) < targetLength {
		return sentence + gofakeit.LetterN(uint(targetLength-len(sentence)))
	}
	return sentence
}

// generateIntegerValue generates an integer with min/max and exclusive bounds.
func (c *Connector) generateIntegerValue(fieldMap map[string]any) (int, error) {
	min := -1000.0
	max := 1000.0

	// Extract minimum
	if minVal, ok := fieldMap["minimum"].(float64); ok {
		min = minVal
	}
	// Check for exclusive minimum
	if exclusive, ok := fieldMap["exclusiveMinimum"].(bool); ok && exclusive {
		min = min + 1
	}

	// Extract maximum
	if maxVal, ok := fieldMap["maximum"].(float64); ok {
		max = maxVal
	}
	// Check for exclusive maximum
	if exclusive, ok := fieldMap["exclusiveMaximum"].(bool); ok && exclusive {
		max = max - 1
	}

	// Check for multipleOf constraint and adjust bounds accordingly
	if multipleOf, ok := fieldMap["multipleOf"].(float64); ok && multipleOf > 0 {
		multiple := int(multipleOf)
		// Validate that multipleOf is a valid non-zero integer
		if multiple > 0 && float64(multiple) == multipleOf {
			// Compute the smallest multiple of multipleOf that is >= minimum
			intMin := int(min)
			if intMin%multiple != 0 {
				if intMin >= 0 {
					// For positive numbers, round up
					intMin = ((intMin / multiple) + 1) * multiple
				} else {
					// For negative numbers, round towards zero (ceiling for negatives)
					intMin = (intMin / multiple) * multiple
				}
			}

			// Compute the largest multiple of multipleOf that is <= maximum
			intMax := int(max)
			if intMax%multiple != 0 {
				if intMax >= 0 {
					// For positive numbers, round down (floor)
					intMax = (intMax / multiple) * multiple
				} else {
					// For negative numbers, round away from zero (floor for negatives)
					intMax = ((intMax / multiple) - 1) * multiple
				}
			}

			// Use adjusted bounds
			return gofakeit.Number(intMin, intMax), nil
		}
		// If validation fails, skip applying multipleOf adjustment
	}

	return gofakeit.Number(int(min), int(max)), nil
}

// generateNumberValue generates a float with min/max and exclusive bounds.
func (c *Connector) generateNumberValue(fieldMap map[string]any) (float64, error) {
	min := -1000.0
	max := 1000.0

	// Extract minimum
	if minVal, ok := fieldMap["minimum"].(float64); ok {
		min = minVal
	}
	// Check for exclusive minimum
	if exclusive, ok := fieldMap["exclusiveMinimum"].(bool); ok && exclusive {
		min = min + 0.001
	}

	// Extract maximum
	if maxVal, ok := fieldMap["maximum"].(float64); ok {
		max = maxVal
	}
	// Check for exclusive maximum
	if exclusive, ok := fieldMap["exclusiveMaximum"].(bool); ok && exclusive {
		max = max - 0.001
	}

	value := generateNumberInRange(&min, &max)

	// Check for multipleOf constraint
	if multipleOf, ok := fieldMap["multipleOf"].(float64); ok && multipleOf > 0 {
		value = float64(int(value/multipleOf)) * multipleOf
	}

	return value, nil
}

// generateArrayValue generates an array with items schema and size constraints.
func (c *Connector) generateArrayValue(fieldMap map[string]any, depth, maxDepth int) ([]any, error) {
	// Extract size constraints
	minItems := 1
	maxItems := 3
	if mi, ok := fieldMap["minItems"].(float64); ok {
		minItems = int(mi)
	}
	if mi, ok := fieldMap["maxItems"].(float64); ok {
		maxItems = int(mi)
	}
	if maxItems < minItems {
		maxItems = minItems
	}

	// Check if items schema exists
	items, hasItems := fieldMap["items"].(map[string]any)
	uniqueItems, _ := fieldMap["uniqueItems"].(bool)

	// Compute theoretical maximum unique values if uniqueItems is true
	if uniqueItems && hasItems {
		// Check if items have an enum constraint
		if enumValues, hasEnum := items["enum"].([]any); hasEnum {
			maxUniqueValues := len(enumValues)
			if minItems > maxUniqueValues {
				return nil, fmt.Errorf("uniqueItems constraint cannot be satisfied: minItems (%d) exceeds number of enum values (%d)", minItems, maxUniqueValues)
			}
			// Adjust maxItems to not exceed available unique values
			if maxItems > maxUniqueValues {
				maxItems = maxUniqueValues
			}
		}
		// Note: For non-enum types, we can't easily compute max unique values,
		// so we'll rely on runtime detection in the loop below
	}

	// Calculate array size
	arraySize := gofakeit.Number(minItems, maxItems)
	arr := make([]any, arraySize)

	// Track unique values if needed
	uniqueMap := make(map[string]bool)

	for i := 0; i < arraySize; i++ {
		var value any
		var err error

		if hasItems && depth < maxDepth {
			// Generate according to items schema
			value, err = c.generateValueFromSchema(items, depth+1, maxDepth)
			if err != nil {
				value = gofakeit.Word()
			}
		} else if depth >= maxDepth {
			// Hit depth limit, use placeholder
			value = generatePlaceholderValue("string")
		} else {
			// No items schema, use random words
			value = gofakeit.Word()
		}

		// Handle uniqueItems constraint
		if uniqueItems {
			valueStr := fmt.Sprintf("%v", value)
			attempts := 0
			for uniqueMap[valueStr] && attempts < 10 {
				if hasItems && depth < maxDepth {
					value, _ = c.generateValueFromSchema(items, depth+1, maxDepth)
				} else {
					value = gofakeit.Word()
				}
				valueStr = fmt.Sprintf("%v", value)
				attempts++
			}
			// If we exhausted retries without finding a unique value, return error
			if uniqueMap[valueStr] {
				return nil, fmt.Errorf("failed to generate unique value for array element %d after %d attempts", i, attempts)
			}
			uniqueMap[valueStr] = true
		}

		arr[i] = value
	}

	return arr, nil
}

// generateObjectValue generates a nested object with properties schema.
func (c *Connector) generateObjectValue(fieldMap map[string]any, depth, maxDepth int) (map[string]any, error) {
	// Check if we've hit depth limit
	if depth >= maxDepth {
		return map[string]any{
			"key": generatePlaceholderValue("string"),
		}, nil
	}

	// Check if properties schema exists
	properties, hasProperties := fieldMap["properties"].(map[string]any)
	if !hasProperties {
		// No properties schema, generate simple object
		return map[string]any{
			"key": gofakeit.Word(),
		}, nil
	}

	// Extract required fields
	requiredFields := make(map[string]bool)
	if reqArray, ok := fieldMap["required"].([]any); ok {
		for _, req := range reqArray {
			if reqStr, ok := req.(string); ok {
				requiredFields[reqStr] = true
			}
		}
	}

	return c.generateNestedObject(properties, requiredFields, depth+1, maxDepth)
}

// generateNestedObject recursively generates a nested object from properties schema.
func (c *Connector) generateNestedObject(properties map[string]any, required map[string]bool, depth, maxDepth int) (map[string]any, error) {
	obj := make(map[string]any)

	for propName, propDef := range properties {
		propMap, ok := propDef.(map[string]any)
		if !ok {
			continue
		}

		// Generate all fields (both required and optional) for better validation success rate
		// Randomly omitting optional fields can trigger edge cases in the jsonschema library
		// that result in template error messages like "{property} does not match schema"

		value, err := c.generateValueFromSchema(propMap, depth, maxDepth)
		if err != nil {
			// Log error with property name and schema details
			propSchemaJSON, _ := json.Marshal(propMap)
			slog.Error("nested property value generation failed, using fallback",
				"propertyName", propName,
				"error", err,
				"propertySchema", string(propSchemaJSON),
				"depth", depth,
				"required", required[propName],
			)
			obj[propName] = gofakeit.Word()
			continue
		}
		obj[propName] = value
	}

	return obj, nil
}

// generateValueFromSchema generates a value from a schema definition.
func (c *Connector) generateValueFromSchema(schema map[string]any, depth, maxDepth int) (any, error) {
	// Check depth limit
	if depth >= maxDepth {
		fieldType, _ := schema["type"].(string)
		return generatePlaceholderValue(fieldType), nil
	}

	return c.generateFieldValue(schema, depth, maxDepth)
}

// generatePlaceholderValue generates a simple placeholder value when depth limit is reached.
func generatePlaceholderValue(fieldType string) any {
	switch fieldType {
	case "string":
		return "placeholder"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return false
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	default:
		return "placeholder"
	}
}

// generateStringByFormat generates a string value based on the format hint.
func generateStringByFormat(format string, minLength, maxLength *int) string {
	var value string

	switch format {
	case "email":
		value = gofakeit.Email()
	case "date-time":
		value = gofakeit.Date().Format(time.RFC3339)
	case "uuid":
		value = uuid.New().String()
	case "uri", "url":
		value = gofakeit.URL()
	case "date":
		value = gofakeit.Date().Format("2006-01-02")
	case "time":
		value = gofakeit.Date().Format("15:04:05")
	case "phone", "telephone":
		value = gofakeit.Phone()
	case "ipv4":
		value = gofakeit.IPv4Address()
	case "ipv6":
		value = gofakeit.IPv6Address()
	case "hostname":
		value = gofakeit.DomainName()
	case "color":
		value = gofakeit.Color()
	case "country":
		value = gofakeit.Country()
	case "city":
		value = gofakeit.City()
	case "street-address":
		value = gofakeit.Street()
	case "postal-code":
		value = gofakeit.Zip()
	default:
		// Randomly choose between different string types
		switch gofakeit.Number(0, 2) {
		case 0:
			value = gofakeit.Name()
		case 1:
			value = gofakeit.Word()
		default:
			value = gofakeit.Sentence(5)
		}
	}

	// Apply length constraints if specified
	if minLength != nil && len(value) < *minLength {
		// Pad with letters to meet minimum length
		value = value + gofakeit.LetterN(uint(*minLength-len(value)))
	}
	if maxLength != nil && len(value) > *maxLength {
		// Truncate to meet maximum length
		value = value[:*maxLength]
	}

	return value
}

// generateNumberInRange generates a number within the specified range.
func generateNumberInRange(min, max *float64) float64 {
	minVal := 0.0
	maxVal := 1000.0

	if min != nil {
		minVal = *min
	}
	if max != nil {
		maxVal = *max
	}

	return gofakeit.Float64Range(minVal, maxVal)
}
