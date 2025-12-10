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

// JSON Schema type constants.
const (
	typeString  = "string"
	typeInteger = "integer"
	typeNumber  = "number"
	typeBoolean = "boolean"
	typeArray   = "array"
	typeObject  = "object"
)

// Generation constants.
const (
	maxDepth                 = 5
	maxRetries               = 100
	maxAttempts              = 3
	defaultPageSize          = 100
	defaultMinStringLength   = 1
	defaultMaxStringLength   = 50
	shortStringThreshold     = 20
	sentenceWordCount        = 5
	defaultMinNumber         = -1000.0
	defaultMaxNumber         = 1000.0
	defaultMinItems          = 1
	defaultMaxItems          = 3
	exclusiveAdjustmentInt   = 1
	exclusiveAdjustmentFloat = 0.001
	uniqueRetryLimit         = 10
	stringFormatChoices      = 2
	splitLimit               = 2
)

// Connector is an in-memory mock connector with JSON schema validation.
type Connector struct {
	client  *common.JSONHTTPClient
	params  *parameters
	schemas SchemaRegistry
	storage Storage
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
//
//nolint:cyclop // Complexity inherent to initialization logic with multiple configuration paths
func NewConnector(opts ...Option) (*Connector, error) {
	// Apply options without pre-populated schemas/storage
	params, err := paramsbuilder.Apply(parameters{}, opts, WithClient(http.DefaultClient))
	if err != nil {
		return nil, err
	}

	// Determine which schemas to use (raw vs struct-derived)
	finalSchemas, err := selectSchemas(params.rawSchemas, params.structSchemas, params.schemas)
	if err != nil {
		return nil, err
	}

	// Parse schemas
	parsedSchemas, err := ParseSchemas(finalSchemas)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schemas: %w", err)
	}

	// Extract special fields from raw schemas before compilation
	// (compilation may not preserve custom x-amp extensions)
	idFields, updatedFields := extractSpecialFields(finalSchemas)

	var store Storage

	// Initialize storage with parsed schemas and special fields
	switch {
	case params.storage != nil:
		store = params.storage
	case params.storageFactory != nil:
		store, err = params.storageFactory(parsedSchemas, idFields, updatedFields, params.observers)
		if err != nil {
			return nil, err
		}
	default:
		store = NewStorage(parsedSchemas, idFields, updatedFields, params.observers)
	}

	return &Connector{
		client: &common.JSONHTTPClient{
			HTTPClient: params.Caller,
		},
		params:  params,
		schemas: parsedSchemas,
		storage: store,
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
//
//nolint:cyclop,funlen // Complexity from pagination, filtering, field selection logic
func (c *Connector) Read(_ context.Context, params common.ReadParams) (*common.ReadResult, error) {
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

	// Parse pagination parameters
	offset := 0

	if params.NextPage != "" {
		var err error

		offset, err = strconv.Atoi(string(params.NextPage))
		if err != nil {
			offset = 0
		}
	}

	pageSize := defaultPageSize
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
	for index, record := range pageRecords {
		// Build fields map with lowercase keys
		// If specific fields are requested, only include those fields
		// Otherwise, include all fields from the record
		fields := make(map[string]any)

		if len(params.Fields) > 0 {
			// Field filtering: only include requested fields
			for field := range params.Fields {
				if value, exists := record[field]; exists {
					fields[strings.ToLower(field)] = value
				}
			}
		} else {
			// No field filtering: include all fields
			for key, value := range record {
				fields[strings.ToLower(key)] = value
			}
		}

		rows[index] = common.ReadResultRow{
			Fields: fields,
			Raw:    record, // Always include full record in Raw
		}
	}

	// Calculate next page token
	var (
		nextPage common.NextPageToken
		done     = true
	)

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
//
//nolint:cyclop,funlen,nestif // Complexity from create/update branching and ID/timestamp generation logic
func (c *Connector) Write(_ context.Context, params common.WriteParams) (*common.WriteResult, error) {
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

	var (
		recordID    string
		finalRecord map[string]any
	)

	// Determine operation (create vs update)
	if params.RecordId == "" {
		// CREATE operation
		idField := c.storage.GetIdFields()[ObjectName(params.ObjectName)]
		updatedField := c.storage.GetUpdatedFields()[ObjectName(params.ObjectName)]

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
		//nolint:modernize // Manual merge to preserve existing fields not in recordMap
		for key, value := range recordMap {
			existing[key] = value
		}

		// Update timestamp field if not explicitly provided in the update
		updatedField := c.storage.GetUpdatedFields()[ObjectName(params.ObjectName)]
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
func (c *Connector) Delete(_ context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
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
func (c *Connector) ListObjectMetadata(
	_ context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
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
			result.AppendError(objectName, fmt.Errorf("%w for %s", ErrSchemaConversion, objectName))

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
//
//nolint:cyclop,funlen // Complexity from validation retry logic and field generation for all property types
func (c *Connector) generateRandomRecordWithDepth(objectName string, depth int) (map[string]any, error) {
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
	idField := c.storage.GetIdFields()[ObjectName(objectName)]
	updatedField := c.storage.GetUpdatedFields()[ObjectName(objectName)]

	// Retry logic for validation failures
	var (
		record        map[string]any
		validationErr error
	)

	for attempt := range maxRetries {
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
				fieldSchemaJSON, marshalErr := json.Marshal(fieldMap)
				if marshalErr != nil {
					fieldSchemaJSON = []byte("{}")
				}

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
	case typeString:
		return c.generateStringValue(fieldMap, format), nil
	case typeInteger:
		return c.generateIntegerValue(fieldMap), nil
	case typeNumber:
		return c.generateNumberValue(fieldMap), nil
	case typeBoolean:
		return gofakeit.Bool(), nil
	case typeArray:
		return c.generateArrayValue(fieldMap, depth, maxDepth)
	case typeObject:
		return c.generateObjectValue(fieldMap, depth, maxDepth)
	default:
		return gofakeit.Word(), nil
	}
}

// generateStringValue generates a string value with pattern and length constraints.
func (c *Connector) generateStringValue(fieldMap map[string]any, format string) string {
	// Check for pattern
	if pattern, ok := fieldMap["pattern"].(string); ok && pattern != "" {
		// Use regex pattern to generate matching string
		for attempt := range maxAttempts {
			value := gofakeit.Regex(pattern)
			if c.validateStringConstraints(value, fieldMap) {
				return value
			}

			_ = attempt // Use the variable to satisfy linter
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
	for attempt := range maxAttempts {
		value := generateStringByFormat(format, minLength, maxLength)
		if c.validateStringConstraints(value, fieldMap) {
			return value
		}

		_ = attempt // Use the variable to satisfy linter
	}

	// Fallback: generate string with length constraints
	return c.generateStringWithLength(minLength, maxLength)
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
	minLen := defaultMinStringLength
	maxLen := defaultMaxStringLength

	if minLength != nil && *minLength > 0 {
		minLen = *minLength
	}

	if maxLength != nil && *maxLength > 0 {
		maxLen = *maxLength
	}

	if maxLen < minLen {
		maxLen = minLen
	}

	// Generate target length
	targetLength := gofakeit.Number(minLen, maxLen)

	// Generate string of appropriate length
	if targetLength <= shortStringThreshold {
		return gofakeit.LetterN(uint(targetLength))
	}

	// For longer strings, use sentence and trim/pad as needed
	sentence := gofakeit.Sentence(targetLength / sentenceWordCount)
	if len(sentence) > targetLength {
		return sentence[:targetLength]
	}

	if len(sentence) < targetLength {
		return sentence + gofakeit.LetterN(uint(targetLength-len(sentence)))
	}

	return sentence
}

// generateIntegerValue generates an integer with min/max and exclusive bounds.
//
//nolint:cyclop,nestif // Complexity from multipleOf constraint handling with positive/negative edge cases
func (c *Connector) generateIntegerValue(fieldMap map[string]any) int {
	minValue := defaultMinNumber
	maxValue := defaultMaxNumber

	// Extract minimum
	if minVal, ok := fieldMap["minimum"].(float64); ok {
		minValue = minVal
	}

	// Check for exclusive minimum
	if exclusive, ok := fieldMap["exclusiveMinimum"].(bool); ok && exclusive {
		minValue += exclusiveAdjustmentInt
	}

	// Extract maximum
	if maxVal, ok := fieldMap["maximum"].(float64); ok {
		maxValue = maxVal
	}

	// Check for exclusive maximum
	if exclusive, ok := fieldMap["exclusiveMaximum"].(bool); ok && exclusive {
		maxValue -= exclusiveAdjustmentInt
	}

	// Check for multipleOf constraint and adjust bounds accordingly
	if multipleOf, ok := fieldMap["multipleOf"].(float64); ok && multipleOf > 0 {
		multiple := int(multipleOf)
		// Validate that multipleOf is a valid non-zero integer
		if multiple > 0 && float64(multiple) == multipleOf {
			// Compute the smallest multiple of multipleOf that is >= minimum
			intMin := int(minValue)
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
			intMax := int(maxValue)
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
			return gofakeit.Number(intMin, intMax)
		}
		// If validation fails, skip applying multipleOf adjustment
	}

	return gofakeit.Number(int(minValue), int(maxValue))
}

// generateNumberValue generates a float with min/max and exclusive bounds.
func (c *Connector) generateNumberValue(fieldMap map[string]any) float64 {
	minValue := defaultMinNumber
	maxValue := defaultMaxNumber

	// Extract minimum
	if minVal, ok := fieldMap["minimum"].(float64); ok {
		minValue = minVal
	}

	// Check for exclusive minimum
	if exclusive, ok := fieldMap["exclusiveMinimum"].(bool); ok && exclusive {
		minValue += exclusiveAdjustmentFloat
	}

	// Extract maximum
	if maxVal, ok := fieldMap["maximum"].(float64); ok {
		maxValue = maxVal
	}

	// Check for exclusive maximum
	if exclusive, ok := fieldMap["exclusiveMaximum"].(bool); ok && exclusive {
		maxValue -= exclusiveAdjustmentFloat
	}

	value := generateNumberInRange(&minValue, &maxValue)

	// Check for multipleOf constraint
	if multipleOf, ok := fieldMap["multipleOf"].(float64); ok && multipleOf > 0 {
		value = float64(int(value/multipleOf)) * multipleOf
	}

	return value
}

// generateArrayValue generates an array with items schema and size constraints.
//
//nolint:cyclop,funlen,gocognit,nestif // Complex logic for uniqueItems, enums, and recursion
func (c *Connector) generateArrayValue(fieldMap map[string]any, depth, maxDepth int) ([]any, error) {
	// Extract size constraints
	minItems := defaultMinItems
	maxItems := defaultMaxItems

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
				return nil, fmt.Errorf("%w: minItems (%d) exceeds number of enum values (%d)",
					ErrUniqueConstraint, minItems, maxUniqueValues)
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

	//nolint:intrange,modernize,varnamelen // Using index for array assignment
	for i := 0; i < arraySize; i++ {
		var (
			value any
			err   error
		)

		switch {
		case hasItems && depth < maxDepth:
			// Generate according to items schema
			value, err = c.generateValueFromSchema(items, depth+1, maxDepth)
			if err != nil {
				value = gofakeit.Word()
			}
		case depth >= maxDepth:
			// Hit depth limit, use placeholder
			value = generatePlaceholderValue(typeString)
		default:
			// No items schema, use random words
			value = gofakeit.Word()
		}

		// Handle uniqueItems constraint
		if uniqueItems {
			valueStr := fmt.Sprintf("%v", value)
			attempts := 0

			for uniqueMap[valueStr] && attempts < uniqueRetryLimit {
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
				return nil, fmt.Errorf("%w for array element %d after %d attempts",
					ErrUniqueValue, i, attempts)
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
			"key": generatePlaceholderValue(typeString),
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
func (c *Connector) generateNestedObject(
	properties map[string]any,
	required map[string]bool,
	depth,
	maxDepth int,
) (map[string]any, error) {
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
			propSchemaJSON, marshalErr := json.Marshal(propMap)
			if marshalErr != nil {
				propSchemaJSON = []byte("{}")
			}

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
	case typeString:
		return "placeholder"
	case typeInteger:
		return 0
	case typeNumber:
		return 0.0
	case typeBoolean:
		return false
	case typeArray:
		return []any{}
	case typeObject:
		return map[string]any{}
	default:
		return "placeholder"
	}
}

// generateStringByFormat generates a string value based on the format hint.
//
//nolint:cyclop,funlen // Complexity from extensive format handling for diverse string types (email, date, UUID, etc.)
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
		switch gofakeit.Number(0, stringFormatChoices) {
		case 0:
			value = gofakeit.Name()
		case 1:
			value = gofakeit.Word()
		default:
			value = gofakeit.Sentence(sentenceWordCount)
		}
	}

	// Apply length constraints if specified
	if minLength != nil && len(value) < *minLength {
		// Pad with letters to meet minimum length
		value += gofakeit.LetterN(uint(*minLength - len(value)))
	}

	if maxLength != nil && len(value) > *maxLength {
		// Truncate to meet maximum length
		value = value[:*maxLength]
	}

	return value
}

// generateNumberInRange generates a number within the specified range.
func generateNumberInRange(minValue, maxValue *float64) float64 {
	minVal := 0.0
	maxVal := defaultMaxNumber

	if minValue != nil {
		minVal = *minValue
	}

	if maxValue != nil {
		maxVal = *maxValue
	}

	return gofakeit.Float64Range(minVal, maxVal)
}

// selectSchemas determines which schemas to use: raw schemas or struct-derived schemas.
// It prioritizes raw schemas if provided, warns if both are provided, and falls back to struct schemas.
func selectSchemas(
	rawSchemas map[string][]byte,
	structSchemas map[string]any,
	schemas map[string]*InputSchema,
) (map[string][]byte, error) {
	switch {
	case len(schemas) > 0:
		out := make(map[string][]byte, len(schemas))

		for name, schema := range schemas {
			bts, err := json.Marshal(schema)
			if err != nil {
				return nil, err
			}

			out[name] = bts
		}

		return out, nil
	case len(rawSchemas) > 0:
		// Raw schemas provided
		// Warn if both raw and struct schemas are provided
		if len(structSchemas) > 0 {
			slog.Warn("both raw schemas and struct schemas provided; using raw schemas",
				"rawSchemaCount", len(rawSchemas),
				"structSchemaCount", len(structSchemas))
		}

		return rawSchemas, nil
	case len(structSchemas) > 0:
		// Derive schemas from structs
		finalSchemas, err := DeriveSchemasFromStructs(structSchemas)
		if err != nil {
			return nil, fmt.Errorf("failed to derive schemas from structs: %w", err)
		}

		return finalSchemas, nil
	default:
		// Neither provided
		return nil, fmt.Errorf("%w: must provide either raw schemas or use WithStructSchemas option", ErrMissingParam)
	}
}

// extractSpecialFields extracts ID and updated timestamp field names from all raw schemas.
// Returns two maps: objectName -> idField and objectName -> updatedField.
func extractSpecialFields(schemas map[string][]byte) (idFields, updatedFields map[string]string) {
	idFields = make(map[string]string)
	updatedFields = make(map[string]string)

	for objectName, rawSchema := range schemas {
		idField, updatedField := extractSpecialFieldsFromRaw(rawSchema)
		if idField != "" {
			idFields[objectName] = idField
		}

		if updatedField != "" {
			updatedFields[objectName] = updatedField
		}
	}

	return idFields, updatedFields
}
