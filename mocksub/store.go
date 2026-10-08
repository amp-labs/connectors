package mocksub

import (
	"maps"
	"sync"

	"github.com/amp-labs/connectors/providers"
)

// Store holds the canned records a mock provider serves from GetRecordsByIds, keyed by object
// name then record id. It is safe for concurrent use so a test can seed records while the flow
// under test reads them.
type Store struct {
	mu      sync.RWMutex
	records map[string]map[string]map[string]any

	// objectNames indexes provider-side object ids to object names, for mock providers
	// mimicking a provider whose events identify their object only by id (e.g. Attio's
	// record.* events carry an id.object_id UUID). Seeded via SeedObjectName and consumed
	// by ObjectIDIndexResolver.
	objectNames map[string]string
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{
		records:     map[string]map[string]map[string]any{},
		objectNames: map[string]string{},
	}
}

// Seed stores a canned record under the given object name and record id, replacing any
// existing record with the same id.
func (s *Store) Seed(objectName, recordID string, record map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.records[objectName] == nil {
		s.records[objectName] = map[string]map[string]any{}
	}

	s.records[objectName][recordID] = maps.Clone(record)
}

// Get returns a copy of the record stored under the object name and record id, and whether
// one exists.
func (s *Store) Get(objectName, recordID string) (map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[objectName][recordID]
	if !ok {
		return nil, false
	}

	return maps.Clone(record), true
}

// SeedObjectName indexes a provider-side object id to its object name (see Store.objectNames).
func (s *Store) SeedObjectName(objectID, objectName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.objectNames[objectID] = objectName
}

// ObjectNameFor returns the object name indexed under the provider-side object id, and whether
// one is seeded.
func (s *Store) ObjectNameFor(objectID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	name, ok := s.objectNames[objectID]

	return name, ok
}

// Clear removes all seeded records and object-name index entries.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = map[string]map[string]map[string]any{}
	s.objectNames = map[string]string{}
}

// stores holds the per-provider singletons handed out by StoreFor.
//
//nolint:gochecknoglobals
var stores = struct {
	mu sync.Mutex
	m  map[providers.Provider]*Store
}{m: map[providers.Provider]*Store{}}

// StoreFor returns the process-wide Store for the given mock provider, creating it on first
// use. Connectors constructed by the connector.New factory read from this store, so tests seed
// records through it before running the flow under test (and Clear it between scenarios).
func StoreFor(provider providers.Provider) *Store {
	stores.mu.Lock()
	defer stores.mu.Unlock()

	store, ok := stores.m[provider]
	if !ok {
		store = NewStore()
		stores.m[provider] = store
	}

	return store
}
