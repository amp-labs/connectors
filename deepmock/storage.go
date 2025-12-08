package deepmock

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kaptinlin/jsonschema"
)

// Storage provides thread-safe in-memory storage for records.
type Storage struct {
	mu            sync.RWMutex
	data          map[string]map[string]map[string]any // objectName -> recordID -> record
	idFields      map[string]string                    // objectName -> ID field name
	updatedFields map[string]string                    // objectName -> updated timestamp field name
}

// NewStorage creates a new Storage instance.
func NewStorage(schemas schemaRegistry, idFields, updatedFields map[string]string) *Storage {
	storage := &Storage{
		data:          make(map[string]map[string]map[string]any),
		idFields:      idFields,
		updatedFields: updatedFields,
	}

	// Initialize object maps
	for objectName := range schemas {
		storage.data[objectName] = make(map[string]map[string]any)
	}

	return storage
}

var errNilRecord = errors.New("record is nil")

// deepCopyRecord creates an independent copy of a record.
func deepCopyRecord(record map[string]any) (map[string]any, error) {
	if record == nil {
		return nil, errNilRecord
	}

	// Use JSON marshal/unmarshal for deep copy
	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	var recordCopy map[string]any
	if err := json.Unmarshal(data, &recordCopy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return recordCopy, nil
}

// deepCopyRecords creates independent copies of a slice of records.
func deepCopyRecords(records []map[string]any) ([]map[string]any, error) {
	if records == nil {
		return nil, errNilRecord
	}

	copies := make([]map[string]any, len(records))

	for i, record := range records {
		recordCopy, err := deepCopyRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to copy record at index %d: %w", i, err)
		}

		copies[i] = recordCopy
	}

	return copies, nil
}

// Store stores a record with the given ID.
func (s *Storage) Store(objectName, recordID string, record map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy before storing
	recordCopy, err := deepCopyRecord(record)
	if err != nil {
		return fmt.Errorf("failed to copy record: %w", err)
	}

	// Initialize object map if needed
	if _, exists := s.data[objectName]; !exists {
		s.data[objectName] = make(map[string]map[string]any)
	}

	s.data[objectName][recordID] = recordCopy

	return nil
}

// Get retrieves a record by ID.
func (s *Storage) Get(objectName, recordID string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objectData, exists := s.data[objectName]
	if !exists {
		return nil, fmt.Errorf("%w: object %s", ErrRecordNotFound, objectName)
	}

	record, exists := objectData[recordID]
	if !exists {
		return nil, fmt.Errorf("%w: record %s", ErrRecordNotFound, recordID)
	}

	// Deep copy before returning
	recordCopy, err := deepCopyRecord(record)
	if err != nil {
		return nil, fmt.Errorf("failed to copy record: %w", err)
	}

	return recordCopy, nil
}

// GetAll retrieves all records for an object.
func (s *Storage) GetAll(objectName string) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objectData, exists := s.data[objectName]
	if !exists {
		return []map[string]any{}, nil
	}

	records := make([]map[string]any, 0, len(objectData))
	for _, record := range objectData {
		records = append(records, record)
	}

	// Deep copy all records
	copies, err := deepCopyRecords(records)
	if err != nil {
		return nil, fmt.Errorf("failed to copy records: %w", err)
	}

	return copies, nil
}

// Delete removes a record by ID.
func (s *Storage) Delete(objectName, recordID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	objectData, exists := s.data[objectName]
	if !exists {
		return fmt.Errorf("%w: object %s", ErrRecordNotFound, objectName)
	}

	if _, exists := objectData[recordID]; !exists {
		return fmt.Errorf("%w: record %s", ErrRecordNotFound, recordID)
	}

	delete(objectData, recordID)

	return nil
}

// List retrieves records filtered by time range.
//
//nolint:cyclop,funlen,gocognit // Complex timestamp parsing and time range filtering
func (s *Storage) List(objectName string, since, until time.Time) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objectData, exists := s.data[objectName]
	if !exists {
		return []map[string]any{}, nil
	}

	// Get the updated field name for this object
	updatedField := s.updatedFields[objectName]

	records := make([]map[string]any, 0)

	for _, record := range objectData {
		// If no updated field, include all records
		if updatedField == "" {
			records = append(records, record)

			continue
		}

		// Filter by timestamp if updated field exists
		//nolint:nestif // Complexity from timestamp type checking and time range filtering
		if updatedValue, exists := record[updatedField]; exists {
			var recordTime time.Time

			hasValidTimestamp := false

			switch updatedVal := updatedValue.(type) {
			case string:
				// Try parsing as RFC3339
				if parsedTime, err := time.Parse(time.RFC3339, updatedVal); err == nil {
					recordTime = parsedTime
					hasValidTimestamp = true
				}
			case int64:
				// Unix timestamp
				recordTime = time.Unix(updatedVal, 0)
				hasValidTimestamp = true
			case int:
				// Unix timestamp as int
				recordTime = time.Unix(int64(updatedVal), 0)
				hasValidTimestamp = true
			case float64:
				// Unix timestamp as float
				recordTime = time.Unix(int64(updatedVal), 0)
				hasValidTimestamp = true
			case json.Number:
				// Handle json.Number type
				if intVal, err := updatedVal.Int64(); err == nil {
					recordTime = time.Unix(intVal, 0)
					hasValidTimestamp = true
				}
			}

			// If time filtering is active and timestamp is invalid, skip the record
			if (!since.IsZero() || !until.IsZero()) && !hasValidTimestamp {
				continue
			}

			// Apply time range filtering only if we have a valid timestamp
			if hasValidTimestamp {
				if !since.IsZero() && recordTime.Before(since) {
					continue
				}

				if !until.IsZero() && recordTime.After(until) {
					continue
				}
			}
		} else if !since.IsZero() || !until.IsZero() {
			// If updated field doesn't exist and time filtering is active, skip the record
			continue
		}

		records = append(records, record)
	}

	// Deep copy filtered records
	copies, err := deepCopyRecords(records)
	if err != nil {
		return nil, fmt.Errorf("failed to copy records: %w", err)
	}

	return copies, nil
}

// generateID generates an ID based on the schema's ID field type.
//
// IMPORTANT: This function must be used in the Write logic to auto-generate IDs for
// create operations. When implementing the connector's Write method:
//   - Detect create operations by checking if the RecordId is missing in the incoming record
//   - Before calling Storage.Store, compute an ID using generateID with the correct schema and idField
//   - The idField should be looked up from storage.idFields[objectName]
//   - Assign the generated ID to the record at the field named by idField
//
// This ensures that auto-generated IDs are always populated according to the schema metadata,
// making the behavior observable to API consumers.
//
//nolint:cyclop // Complexity from schema marshaling/unmarshaling and nested map traversal to extract type
func generateID(schema *jsonschema.Schema, idField string) any {
	if schema == nil || idField == "" {
		return uuid.New().String()
	}

	// Extract schema as map to check field type
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return uuid.New().String()
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		return uuid.New().String()
	}

	properties, hasProperties := schemaMap["properties"].(map[string]any)
	if !hasProperties {
		return uuid.New().String()
	}

	fieldDef, hasFieldDef := properties[idField].(map[string]any)
	if !hasFieldDef {
		return uuid.New().String()
	}

	fieldType, hasType := fieldDef["type"].(string)
	if !hasType {
		return uuid.New().String()
	}

	switch fieldType {
	case typeInteger:
		return time.Now().UnixNano()
	case typeString:
		return uuid.New().String()
	default:
		return uuid.New().String()
	}
}

// generateTimestamp generates a timestamp based on the schema's updated field type.
//
// IMPORTANT: This function must be used in the Write logic to auto-generate timestamps for
// both create and update operations. When implementing the connector's Write method:
//   - For both create and update operations, compute an updated timestamp using generateTimestamp
//   - The updatedField should be looked up from storage.updatedFields[objectName]
//   - Assign the generated timestamp to the record at the field named by updatedField
//   - This should happen before calling Storage.Store
//
// This ensures that the updated timestamp is always current and properly formatted according to
// the schema metadata, making the auto-generation behavior observable to API consumers.
//
//nolint:cyclop // Complexity from schema marshaling/unmarshaling and nested map traversal to extract type
func generateTimestamp(schema *jsonschema.Schema, updatedField string) any {
	if schema == nil || updatedField == "" {
		return time.Now().Unix()
	}

	// Extract schema as map to check field type
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return time.Now().Unix()
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		return time.Now().Unix()
	}

	properties, hasProperties := schemaMap["properties"].(map[string]any)
	if !hasProperties {
		return time.Now().Unix()
	}

	fieldDef, hasFieldDef := properties[updatedField].(map[string]any)
	if !hasFieldDef {
		return time.Now().Unix()
	}

	fieldType, hasType := fieldDef["type"].(string)
	if !hasType {
		return time.Now().Unix()
	}

	switch fieldType {
	case typeInteger:
		return time.Now().Unix()
	case typeString:
		return time.Now().Format(time.RFC3339)
	default:
		return time.Now().Unix()
	}
}
