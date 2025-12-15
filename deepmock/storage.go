package deepmock

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/future"
	"github.com/google/uuid"
	"github.com/kaptinlin/jsonschema"
)

// Storage defines the interface for persisting and retrieving mock API records.
// It provides a flexible abstraction layer that allows different storage backends
// (in-memory, file-based, database-backed) to be used with the deepmock connector.
//
// All implementations must be thread-safe and support concurrent reads and writes.
// Records are organized by object name (e.g., "contact", "account") and identified
// by unique record IDs within each object type.
//
// The interface supports CRUD operations (Store, Get, Delete) as well as bulk
// retrieval (GetAll, List) with optional time-based filtering for incremental
// sync scenarios.
type Storage interface {
	// Store persists a record with the specified ID for the given object type.
	// If a record with the same objectName and recordID already exists, it will be replaced.
	//
	// Parameters:
	//   - objectName: The type of object being stored (e.g., "contact", "account")
	//   - recordID: A unique identifier for this record within the object type
	//   - record: The record data as a map of field names to values
	//
	// Returns an error if:
	//   - The record is nil
	//   - The storage operation fails
	//
	// Implementations should deep copy the record to prevent external modifications
	// from affecting stored data.
	Store(objectName, recordID string, record map[string]any) error

	// Get retrieves a single record by its ID for the specified object type.
	//
	// Parameters:
	//   - objectName: The type of object to retrieve (e.g., "contact", "account")
	//   - recordID: The unique identifier of the record to retrieve
	//
	// Returns:
	//   - The record data as a map of field names to values
	//   - An error if the object type or record ID does not exist (ErrRecordNotFound)
	//
	// Implementations should return a deep copy of the record to prevent external
	// modifications from affecting stored data.
	Get(objectName, recordID string) (map[string]any, error)

	// GetAll retrieves all records for the specified object type.
	//
	// Parameters:
	//   - objectName: The type of object to retrieve (e.g., "contact", "account")
	//
	// Returns:
	//   - A slice of all records for this object type (empty slice if none exist)
	//   - An error if the retrieval operation fails
	//
	// Unlike Get, this method returns an empty slice rather than an error when
	// the object type has no records. Implementations should return deep copies
	// of all records.
	GetAll(objectName string) ([]map[string]any, error)

	// Delete removes a record by its ID for the specified object type.
	//
	// Parameters:
	//   - objectName: The type of object containing the record (e.g., "contact", "account")
	//   - recordID: The unique identifier of the record to delete
	//
	// Returns an error if:
	//   - The object type does not exist (ErrRecordNotFound)
	//   - The record ID does not exist (ErrRecordNotFound)
	//   - The delete operation fails
	Delete(objectName, recordID string) error

	// List retrieves records for the specified object type, optionally filtered by
	// a time range based on the object's updated timestamp field.
	//
	// This method is designed to support incremental sync scenarios where only records
	// modified within a specific time window need to be retrieved.
	//
	// Parameters:
	//   - objectName: The type of object to retrieve (e.g., "contact", "account")
	//   - since: Start of time range (inclusive). Zero value means no lower bound.
	//   - until: End of time range (inclusive). Zero value means no upper bound.
	//
	// Returns:
	//   - A slice of records matching the time range criteria (empty slice if none match)
	//   - An error if the retrieval operation fails
	//
	// Time filtering behavior:
	//   - If both since and until are zero, all records are returned (equivalent to GetAll)
	//   - If an updated timestamp field is not configured for this object type, all records
	//     are returned regardless of time range parameters
	//   - Records with missing or unparseable timestamp values are excluded when time
	//     filtering is active
	//   - Supported timestamp formats: RFC3339 strings, Unix timestamps (int/int64/float64)
	List(objectName string, since, until time.Time) ([]map[string]any, error)

	// GetIdFields returns a mapping of object names to their ID field names.
	//
	// The ID field is the field within each record that serves as its unique identifier.
	// For example, a "contact" object might use "id" or "contact_id" as its ID field.
	//
	// Returns:
	//   - A map where keys are object names and values are the corresponding ID field names
	//
	// This mapping is used by the connector to:
	//   - Extract record IDs when processing read responses
	//   - Generate appropriate IDs for create operations
	//   - Validate that records contain the required ID field
	//
	// Implementations should return a copy of the internal mapping to prevent external
	// modifications.
	GetIdFields() map[ObjectName]string

	// GetUpdatedFields returns a mapping of object names to their updated timestamp field names.
	//
	// The updated field is the field within each record that tracks when it was last modified.
	// For example, a "contact" object might use "updated_at", "modified_date", or "last_modified_time".
	//
	// Returns:
	//   - A map where keys are object names and values are the corresponding updated field names
	//
	// This mapping is used by the connector to:
	//   - Filter records by modification time in List operations
	//   - Auto-generate timestamps for create and update operations
	//   - Support incremental sync scenarios
	//
	// Implementations should return a copy of the internal mapping to prevent external
	// modifications.
	GetUpdatedFields() map[ObjectName]string
}

// storage provides thread-safe in-memory storage for records.
type storage struct {
	mu            sync.RWMutex
	data          map[ObjectName]map[RecordID]common.Record // objectName -> recordID -> record
	idFields      map[ObjectName]string                     // objectName -> ID field name
	updatedFields map[ObjectName]string                     // objectName -> updated timestamp field name
	observers     []func(action string, record map[string]any)
}

// Compile-time check to ensure storage implements the Storage interface.
var _ Storage = (*storage)(nil)

// NewStorage creates a new Storage instance.
func NewStorage(
	schemas SchemaRegistry,
	idFields, updatedFields map[string]string,
	observers []func(action string, record map[string]any),
) Storage {
	store := &storage{
		data:          make(map[ObjectName]map[RecordID]common.Record),
		idFields:      make(map[ObjectName]string),
		updatedFields: make(map[ObjectName]string),
		observers:     observers,
	}

	// Initialize object maps and convert string keys to typed keys
	for objectName := range schemas {
		store.data[ObjectName(objectName)] = make(map[RecordID]common.Record)
	}

	// Convert string maps to typed maps
	for objectName, fieldName := range idFields {
		store.idFields[ObjectName(objectName)] = fieldName
	}

	for objectName, fieldName := range updatedFields {
		store.updatedFields[ObjectName(objectName)] = fieldName
	}

	return store
}

// errNilRecord is returned when attempting to copy a nil record.
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
func (s *storage) Store(objectName, recordID string, record map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy before storing
	recordCopy, err := deepCopyRecord(record)
	if err != nil {
		return fmt.Errorf("failed to copy record: %w", err)
	}

	// Initialize object map if needed
	objName := ObjectName(objectName)
	if _, exists := s.data[objName]; !exists {
		s.data[objName] = make(map[RecordID]common.Record)
	}

	s.data[objName][RecordID(recordID)] = recordCopy

	// Send to observers, if any
	for _, observe := range s.observers {
		recordCopyCopy, err := deepCopyRecord(recordCopy)
		if err == nil {
			s.sendRecordToObserver("write", recordCopyCopy, observe)
		} else {
			slog.Warn("deepCopyRecord failed", "error", err)

			s.sendRecordToObserver("write", recordCopy, observe)
		}
	}

	return nil
}

// Get retrieves a record by ID.
func (s *storage) Get(objectName, recordID string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objectData, exists := s.data[ObjectName(objectName)]
	if !exists {
		return nil, fmt.Errorf("%w: object %s", ErrRecordNotFound, objectName)
	}

	record, exists := objectData[RecordID(recordID)]
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
func (s *storage) GetAll(objectName string) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objectData, exists := s.data[ObjectName(objectName)]
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
func (s *storage) Delete(objectName, recordID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	objectData, exists := s.data[ObjectName(objectName)]
	if !exists {
		return fmt.Errorf("%w: object %s", ErrRecordNotFound, objectName)
	}

	record, exists := objectData[RecordID(recordID)]
	if !exists {
		return fmt.Errorf("%w: record %s", ErrRecordNotFound, recordID)
	}

	delete(objectData, RecordID(recordID))

	// Send to observers, if any
	for _, observe := range s.observers {
		s.sendRecordToObserver("delete", record, observe)
	}

	return nil
}

// List retrieves records filtered by time range.
//
//nolint:cyclop,funlen,gocognit // Complex timestamp parsing and time range filtering
func (s *storage) List(objectName string, since, until time.Time) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objectData, exists := s.data[ObjectName(objectName)]
	if !exists {
		return []map[string]any{}, nil
	}

	// Get the updated field name for this object
	updatedField := s.updatedFields[ObjectName(objectName)]

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

// GetIdFields returns a copy of the object name to ID field name mapping.
// This ensures external callers cannot modify the internal mapping.
func (s *storage) GetIdFields() map[ObjectName]string {
	out := make(map[ObjectName]string, len(s.idFields))

	maps.Copy(out, s.idFields)

	return out
}

// GetUpdatedFields returns a copy of the object name to updated timestamp field name mapping.
// This ensures external callers cannot modify the internal mapping.
func (s *storage) GetUpdatedFields() map[ObjectName]string {
	out := make(map[ObjectName]string, len(s.updatedFields))

	maps.Copy(out, s.updatedFields)

	return out
}

// sendRecordToObserver asynchronously notifies an observer of a storage action.
// The observer function is invoked in a separate goroutine to prevent blocking
// the storage operation. Any errors from the goroutine are intentionally ignored.
func (s *storage) sendRecordToObserver(
	action string,
	record map[string]any,
	observer func(action string, record map[string]any),
) {
	_ = future.Go[struct{}](func() (struct{}, error) {
		observer(action, record)

		return struct{}{}, nil
	})
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
