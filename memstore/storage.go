package memstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/future"
	"github.com/google/uuid"
	"github.com/kaptinlin/jsonschema"
)

// NotifyCallback is invoked when a storage operation occurs on a subscribed object.
// It receives the subscription context, action type (e.g., "create:account", "update:contact"),
// object name, record ID, and the full record data.
type NotifyCallback func(
	subscription *SubscriptionContext,
	action string,
	objectName string,
	recordID string,
	record map[string]any,
)

// SubscriptionEvents maps object names to the events and fields being watched for that object.
type SubscriptionEvents = map[common.ObjectName]common.ObjectEvents

// SubscriptionContext represents an active subscription to storage events.
// It contains the callback function, subscription configuration, and metadata
// needed to filter and deliver notifications.
type SubscriptionContext struct {
	Notify             NotifyCallback      `json:"-"`
	Id                 string              `json:"id"`
	SubscriptionEvents SubscriptionEvents  `json:"subscriptionEvents"`
	RegistrationRef    string              `json:"registrationRef"`
	RegistrationResult *RegistrationResult `json:"registrationResult"`
	Metadata           map[string]any      `json:"metadata"`
}

// Storage defines the interface for persisting and retrieving mock API records.
// It provides a flexible abstraction layer that allows different storage backends
// (in-memory, file-based, database-backed) to be used with the memstore connector.
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
	//   - action: Optional action type ("create", "update", or "write"). If omitted, defaults to "write".
	//
	// Returns an error if:
	//   - The record is nil
	//   - The storage operation fails
	//
	// Implementations should deep copy the record to prevent external modifications
	// from affecting stored data.
	Store(objectName, recordID string, record map[string]any, action ...string) error

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
	//   - Records with missing or unparsable timestamp values are excluded when time
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

	// GetAssociations returns a mapping of object names to their association metadata.
	//
	// Each object can have zero or more fields configured as associations to other objects.
	// For example, a "contact" object might have an "account_id" field that is a foreign key
	// association to the "account" object.
	//
	// Returns:
	//   - A map where keys are object names and values are maps of field names to association schemas
	//   - Each association schema contains the type (foreignKey, reverseLookup, junction),
	//     target object, and other configuration needed to expand the association
	//
	// This mapping is used by the connector to:
	//   - Expand associations during Read operations by fetching related records
	//   - Validate foreign key references during Write operations
	//   - Include association definitions in object metadata
	//
	// Implementations should return a copy of the internal mapping to prevent external
	// modifications.
	GetAssociations() map[ObjectName]map[string]*AssociationSchema

	// Subscribe registers a new subscription to receive notifications for storage events.
	// The subscription context contains the callback function, event filters, and metadata.
	//
	// Parameters:
	//   - subscription: The subscription context containing callback and event configuration
	//
	// Returns an error if:
	//   - The subscription is nil
	//   - A subscription with the same ID already exists
	//
	// Once subscribed, the callback will be invoked asynchronously for all matching
	// storage operations (Store, Delete) based on the subscription's event filters.
	Subscribe(subscription *SubscriptionContext) error

	// Unsubscribe removes an active subscription by its ID.
	//
	// Parameters:
	//   - id: The unique identifier of the subscription to remove
	//
	// Returns an error if no subscription exists with the given ID.
	//
	// After unsubscribing, the callback will no longer receive notifications.
	Unsubscribe(id string) error
}

// storage provides thread-safe in-memory storage for records.
// It implements the Storage interface and uses a read-write mutex to protect
// concurrent access to its internal maps. Records are organized by object type
// and record ID, and each object type can have custom ID and timestamp field names.
type storage struct {
	mu            sync.RWMutex                                 // Protects concurrent access to all fields
	data          map[ObjectName]map[RecordID]common.Record    // objectName -> recordID -> record
	idFields      map[ObjectName]string                        // objectName -> ID field name
	updatedFields map[ObjectName]string                        // objectName -> updated timestamp field name
	associations  map[ObjectName]map[string]*AssociationSchema // objectName -> fieldName -> association metadata
	subscriptions map[string]*SubscriptionContext              // subscriptionID -> subscription context
}

// Unsubscribe removes an active subscription by its ID.
// After unsubscribing, the callback will no longer receive notifications for storage events.
// Returns ErrObserverNotFound if no subscription exists with the given ID.
func (s *storage) Unsubscribe(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.subscriptions[id]
	if !ok {
		return ErrObserverNotFound
	}

	delete(s.subscriptions, id)

	return nil
}

// Compile-time check to ensure storage implements the Storage interface.
var _ Storage = (*storage)(nil)

// NewStorage creates a new Storage instance with the specified schemas and field mappings.
//
// Parameters:
//   - schemas: Registry of object schemas used to initialize storage maps for each object type
//   - idFields: Mapping of object names to their ID field names (e.g., "contact" -> "id")
//   - updatedFields: Mapping of object names to their timestamp field names (e.g., "contact" -> "updated_at")
//   - associations: Mapping of object names to their field-level association metadata
//     (e.g., "contact" -> "account_id" -> AssociationSchema)
//
// Returns a thread-safe in-memory storage implementation with all object types initialized
// and ready to accept records.
func NewStorage(
	schemas SchemaRegistry,
	idFields map[string]string,
	updatedFields map[string]string,
	associations map[string]map[string]*AssociationSchema,
) Storage {
	store := &storage{
		data:          make(map[ObjectName]map[RecordID]common.Record),
		idFields:      make(map[ObjectName]string),
		updatedFields: make(map[ObjectName]string),
		associations:  make(map[ObjectName]map[string]*AssociationSchema),
		subscriptions: make(map[string]*SubscriptionContext),
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

	// Convert association maps to typed maps and deep copy
	for objectName, fieldAssocs := range associations {
		store.associations[ObjectName(objectName)] = make(map[string]*AssociationSchema, len(fieldAssocs))

		for fieldName, assoc := range fieldAssocs {
			// Deep copy the AssociationSchema
			assocCopy := *assoc
			store.associations[ObjectName(objectName)][fieldName] = &assocCopy
		}
	}

	return store
}

// errNilRecord is returned when attempting to copy a nil record.
var errNilRecord = errors.New("record is nil")

// deepCopyRecord creates an independent copy of a record.
// This ensures that modifications to the returned record do not affect the original,
// and vice versa. The function uses JSON marshaling/unmarshaling to handle nested
// structures and complex types.
//
// Returns errNilRecord if the input record is nil.
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
// Each record in the slice is deep copied to ensure modifications to the returned
// slice or its records do not affect the originals.
//
// Returns errNilRecord if the input slice is nil, or an error if any individual
// record fails to copy.
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

// Store stores a record with the given ID and notifies all active subscriptions.
// The method automatically adds the record ID to the record using the object's configured
// ID field name. If the record already exists, it will be replaced and an "update" event
// will be generated; otherwise, a "create" event will be generated.
//
// The action parameter should be "create", "update", or "write" (defaults to "write" for
// backwards compatibility). The action type is included in the observer notification as
// "action:objectName" (e.g., "create:account", "update:contact").
//
// All subscriptions are notified asynchronously in separate goroutines to prevent blocking.
func (s *storage) Store(objectName, recordID string, record map[string]any, action ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy before storing
	recordCopy, err := deepCopyRecord(record)
	if err != nil {
		return fmt.Errorf("failed to copy record: %w", err)
	}

	// Add the ID to the record (needed for observers and record retrieval)
	// Use the schema's ID field name, defaulting to "id"
	idField := s.idFields[ObjectName(objectName)]
	if idField == "" {
		idField = "id"
	}

	recordCopy[idField] = recordID

	// Initialize object map if needed
	objName := ObjectName(objectName)
	if _, exists := s.data[objName]; !exists {
		s.data[objName] = make(map[RecordID]common.Record)
	}

	var changedFields map[string]struct{}

	previous, hasPrevious := s.data[objName][RecordID(recordID)]

	eventType := common.SubscriptionEventTypeCreate

	if hasPrevious {
		eventType = common.SubscriptionEventTypeUpdate

		changedFields = getChangedFields(previous, recordCopy)
	}

	s.data[objName][RecordID(recordID)] = recordCopy

	// Send to observers, if any
	// Include object name in action format: "action:objectName" (e.g., "create:accounts", "update:contacts")
	actionType := "write" // default for backwards compatibility
	if len(action) > 0 && action[0] != "" {
		actionType = action[0]
	}

	observerAction := actionType + ":" + objectName

	for _, sub := range s.subscriptions {
		recordCopyCopy, err := deepCopyRecord(recordCopy)
		if err == nil {
			s.sendRecordToSubscriber(observerAction, eventType, objectName, recordID, recordCopyCopy, changedFields, sub)
		} else {
			slog.Warn("deepCopyRecord failed", "error", err)

			s.sendRecordToSubscriber(observerAction, eventType, objectName, recordID, recordCopy, changedFields, sub)
		}
	}

	return nil
}

// getChangedFields compares two records and returns the set of fields that differ.
// A field is considered changed if:
//   - It exists in oldRecord but not in newRecord (deletion)
//   - It exists in newRecord but not in oldRecord (addition)
//   - It exists in both but has different values (modification)
//
// The returned map uses empty struct values for memory efficiency.
func getChangedFields(oldRecord, newRecord common.Record) map[string]struct{} {
	fields := make(map[string]struct{})

	for field := range oldRecord {
		oldValue := oldRecord[field]

		newValue, hasField := newRecord[field]

		if !hasField {
			fields[field] = struct{}{}

			continue
		}

		if !reflect.DeepEqual(oldValue, newValue) {
			fields[field] = struct{}{}

			continue
		}
	}

	for field := range newRecord {
		newValue := newRecord[field]

		oldValue, hasField := oldRecord[field]

		if !hasField {
			fields[field] = struct{}{}

			continue
		}

		if !reflect.DeepEqual(oldValue, newValue) {
			fields[field] = struct{}{}

			continue
		}
	}

	return fields
}

// Get retrieves a record by ID and returns a deep copy to prevent external modifications.
// Returns ErrRecordNotFound if either the object type or record ID does not exist.
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

// GetAll retrieves all records for an object and returns deep copies to prevent external modifications.
// Returns an empty slice if the object type doesn't exist or has no records.
// Unlike Get, this method does not return an error for non-existent object types.
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

// Delete removes a record by ID and notifies all active subscriptions.
// Returns ErrRecordNotFound if either the object type or record ID does not exist.
// All subscriptions are notified asynchronously with a "delete:objectName" action.
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
	// Include object name in action format: "delete:objectName"
	action := "delete:" + objectName
	for _, sub := range s.subscriptions {
		s.sendRecordToSubscriber(action, common.SubscriptionEventTypeDelete,
			objectName, recordID, record, nil, sub)
	}

	return nil
}

// List retrieves records filtered by time range and returns deep copies to prevent external modifications.
// Records are filtered based on the object's configured updated timestamp field.
//
// Behavior:
//   - If both since and until are zero, returns all records (equivalent to GetAll)
//   - If no updated field is configured for the object, returns all records regardless of time range
//   - Records with missing or unparsable timestamps are excluded when time filtering is active
//   - Supports RFC3339 strings and Unix timestamps (int, int64, float64, json.Number)
//
// Returns an empty slice if the object type doesn't exist or no records match the time range.
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

// GetAssociations returns a deep copy of the object name to association metadata mapping.
// This ensures external callers cannot modify the internal mapping.
func (s *storage) GetAssociations() map[ObjectName]map[string]*AssociationSchema {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[ObjectName]map[string]*AssociationSchema, len(s.associations))

	for objectName, fieldAssocs := range s.associations {
		out[objectName] = make(map[string]*AssociationSchema, len(fieldAssocs))

		for fieldName, assoc := range fieldAssocs {
			// Deep copy the AssociationSchema to prevent external modifications
			assocCopy := *assoc
			out[objectName][fieldName] = &assocCopy
		}
	}

	return out
}

// Subscribe adds additional observer functions to the storage.
// This allows multiple connector instances sharing the same storage to register
// their own observers. The method is thread-safe.
func (s *storage) Subscribe(subscription *SubscriptionContext) error {
	if subscription == nil {
		return ErrSubscriptionNil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.subscriptions[subscription.Id]
	if ok {
		return fmt.Errorf("%w: %q", ErrSubscriptionExists, subscription.Id)
	}

	s.subscriptions[subscription.Id] = subscription

	return nil
}

// sendRecordToSubscriber asynchronously notifies an observer of a storage action.
// The observer function is invoked in a separate goroutine to prevent blocking
// the storage operation. Any errors from the goroutine are intentionally ignored.
func (s *storage) sendRecordToSubscriber(
	action string,
	evtType common.SubscriptionEventType,
	objectName string,
	recordID string,
	record map[string]any,
	changedFields map[string]struct{},
	subscription *SubscriptionContext,
) {
	_ = future.Go[struct{}](func() (struct{}, error) {
		if wantNotification(evtType, objectName, changedFields, subscription) {
			subscription.Notify(subscription, action, objectName, recordID, record)
		}

		return struct{}{}, nil
	})
}

// wantNotification determines whether a subscription should receive a notification
// for the given event. It checks:
//  1. Whether the subscription is watching the object type
//  2. Whether the subscription is interested in the event type (create/update/delete)
//  3. For update events, whether any watched fields have changed
//
// For create and delete events, field filtering is not applied.
// For update events with WatchFieldsAll=true, all updates are reported.
// For update events with specific WatchFields, only updates to those fields trigger notifications.
func wantNotification(
	evtType common.SubscriptionEventType,
	objectName string,
	changedFields map[string]struct{},
	subscription *SubscriptionContext,
) bool {
	evts, ok := subscription.SubscriptionEvents[common.ObjectName(objectName)]
	if !ok {
		return false
	}

	if !slices.Contains(evts.Events, evtType) {
		return false
	}

	if evtType == common.SubscriptionEventTypeDelete || evtType == common.SubscriptionEventTypeCreate {
		return true
	}

	if evts.WatchFieldsAll {
		return true
	}

	foundRelevantFields := false

	for _, field := range evts.WatchFields {
		_, found := changedFields[field]

		if found {
			foundRelevantFields = true

			break
		}
	}

	return foundRelevantFields
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
