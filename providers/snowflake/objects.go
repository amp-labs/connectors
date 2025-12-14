package snowflake

import (
	"fmt"
	"strings"

	"github.com/amp-labs/amp-common/jsonpath"
)

// Metadata keys for per-object configuration.
// These are the property names expected in JSONPath keys.
//
// Structure:
//
//	$['objects']['objName']['dynamicTable']['query']      - SQL query
//	$['objects']['objName']['dynamicTable']['primaryKey'] - Primary key column
//	$['objects']['objName']['dynamicTable']['timestampColumn'] - Timestamp column
//	$['objects']['objName']['dynamicTable']['targetLag']  - Refresh interval
//	$['objects']['objName']['dynamicTable']['name']       - Generated DT name
//	$['objects']['objName']['stream']['name']             - Generated stream name
//	$['objects']['objName']['stream']['consumptionTable'] - Consumption table
const (
	// MetadataKeyQuery is the SQL query defining the data source (nested under 'dynamicTable').
	MetadataKeyQuery = "query"

	// MetadataKeyDynamicTable is the parent key for dynamic table configuration.
	MetadataKeyDynamicTable = "dynamicTable"
	// MetadataKeyStream is the parent key for stream configuration.
	MetadataKeyStream = "stream"

	// MetadataKeyPrimaryKey is the primary key column (nested under 'dynamicTable').
	MetadataKeyPrimaryKey = "primaryKey"
	// MetadataKeyTimestampColumn is the timestamp column (nested under 'dynamicTable').
	MetadataKeyTimestampColumn = "timestampColumn"
	// MetadataKeyTargetLag is the refresh interval (nested under 'dynamicTable').
	MetadataKeyTargetLag = "targetLag"
	// MetadataKeyName is the generated name (nested under 'dynamicTable' or 'stream').
	MetadataKeyName = "name"
	// MetadataKeyConsumptionTable is the table used for advancing stream offsets (nested under 'stream').
	MetadataKeyConsumptionTable = "consumptionTable"
)

const (
	// objectsKey is the root key for all object configurations.
	objectsKey = "objects"

	// expectedPathDepth is the expected depth for all object configuration paths.
	// All paths follow: $['objects']['objectName']['parent']['property'].
	expectedPathDepth = 4
)

// Objects holds the configuration for multiple Snowflake objects.
// Configuration refers to any parameters that help us create the
// object on the snowflake i.e. the SQL query, stream name, etc.
type Objects map[string]objectConfig

// Validate checks that all objects have the required configuration.
// Required fields: query, dynamicTable.primaryKey
// Returns an error describing all validation failures.
func (s *Objects) Validate() error {
	if s == nil || len(*s) == 0 {
		return nil
	}

	var errors []string

	for objectName, cfg := range *s {
		if cfg.dynamicTable.query == "" {
			errors = append(errors, fmt.Sprintf("object %q: missing required field '%s.%s'",
				objectName, MetadataKeyDynamicTable, MetadataKeyQuery))
		}

		if cfg.dynamicTable.primaryKey == "" {
			errors = append(errors, fmt.Sprintf("object %q: missing required field '%s.%s'",
				objectName, MetadataKeyDynamicTable, MetadataKeyPrimaryKey))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w:\n  - %s", errObjectsValidationFailed, strings.Join(errors, "\n  - "))
	}

	return nil
}

func (s *Objects) ToMetadataMap() map[string]string {
	result := make(map[string]string)

	// Helper to add non-empty values only
	addIfNotEmpty := func(key, value string) {
		if value != "" {
			result[key] = value
		}
	}

	// Reverse of newSnowflakeObjects
	// Structure: $['objects']['objName']['dynamicTable']['property'] for DT
	//            $['objects']['objName']['stream']['property'] for Stream
	for objectName, cfg := range *s {
		// Dynamic Table properties (nested under 'dynamicTable')
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyDynamicTable, MetadataKeyQuery),
			cfg.dynamicTable.query,
		)
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyDynamicTable, MetadataKeyPrimaryKey),
			cfg.dynamicTable.primaryKey,
		)
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyDynamicTable, MetadataKeyTimestampColumn),
			cfg.dynamicTable.timestampColumn,
		)
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyDynamicTable, MetadataKeyTargetLag),
			cfg.dynamicTable.targetLag,
		)
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyDynamicTable, MetadataKeyName),
			cfg.dynamicTable.name,
		)

		// Stream properties (nested under 'stream')
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyStream, MetadataKeyName),
			cfg.stream.name,
		)
		addIfNotEmpty(
			jsonpath.ToNestedPath(objectsKey, objectName, MetadataKeyStream, MetadataKeyConsumptionTable),
			cfg.stream.consumptionTable,
		)
	}

	return result
}

func (s *Objects) Get(objectName string) (*objectConfig, bool) {
	cfg, ok := (*s)[objectName]

	return &cfg, ok
}

// objectConfig holds the configuration for a single Snowflake object.
// Structure mirrors the metadata JSONPath:
//
//	$['objects']['objName']['dynamicTable'][...]
//	$['objects']['objName']['stream'][...]
type objectConfig struct {
	// dynamicTable holds Dynamic Table specific configuration.
	dynamicTable dynamicTableConfig

	// stream holds Stream specific configuration.
	stream streamConfig
}

// dynamicTableConfig holds configuration for a Snowflake Dynamic Table.
type dynamicTableConfig struct {
	// query is the SQL query defining the data that we treat as an object.
	query string

	// primaryKey is the column used for consistent ordering during pagination.
	// This should be a unique, stable column (e.g., "id", "account_id").
	primaryKey string

	// timestampColumn is the column used for incremental filtering.
	timestampColumn string

	// targetLag is the Dynamic Table refresh interval (e.g., "1 hour").
	targetLag string

	// name is the generated Dynamic Table name (set after PostAuth flow).
	// Used to read data when we need to do non-incremental syncs.
	name string
}

// streamConfig holds configuration for a Snowflake Stream.
type streamConfig struct {
	// name is the generated Stream name (set after PostAuth flow).
	// Used to read data for incremental syncs.
	name string

	// consumptionTable is the table used to advance stream offsets.
	// Set after PostAuth flow. Shared across all streams but stored per-object
	// so each object's metadata is self-contained.
	consumptionTable string
}

func newSnowflakeObjects(paramsMap map[string]string) (*Objects, error) {
	result := make(Objects)

	for key, value := range paramsMap {
		if !jsonpath.IsNestedPath(key) {
			continue
		}

		// Parse the JSONPath to extract segments (this also validates the path)
		segments, err := jsonpath.ParsePath(key)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", key, err)
		}

		// Skip paths that don't belong to objects
		if segments[0].Key != objectsKey {
			continue
		}

		// Validate path structure for objects: $['objects']['objectName']['parent']['property']
		if len(segments) != expectedPathDepth {
			return nil, fmt.Errorf(
				"%w for %q: expected %d segments, got %d",
				errInvalidPathDepth, key, expectedPathDepth, len(segments),
			)
		}

		// Extract path components
		objectName := segments[1].Key
		parent := segments[2].Key
		property := segments[3].Key

		cfg := result[objectName]

		// Set property based on parent type
		var setErr error

		switch parent {
		case MetadataKeyDynamicTable:
			cfg, setErr = setDynamicTableProperty(cfg, property, value)
		case MetadataKeyStream:
			cfg, setErr = setStreamProperty(cfg, property, value)
		default:
			return nil, fmt.Errorf(
				"%w %q in path %q: must be %q or %q",
				errInvalidParentKey, parent, key, MetadataKeyDynamicTable, MetadataKeyStream,
			)
		}

		if setErr != nil {
			return nil, fmt.Errorf("%w in path %q: %w", errUnknownProperty, key, setErr)
		}

		result[objectName] = cfg
	}

	return &result, nil
}

func setDynamicTableProperty(cfg objectConfig, property, value string) (objectConfig, error) {
	switch property {
	case MetadataKeyQuery:
		cfg.dynamicTable.query = value
	case MetadataKeyPrimaryKey:
		cfg.dynamicTable.primaryKey = value
	case MetadataKeyTimestampColumn:
		cfg.dynamicTable.timestampColumn = value
	case MetadataKeyTargetLag:
		cfg.dynamicTable.targetLag = value
	case MetadataKeyName:
		cfg.dynamicTable.name = value
	default:
		return cfg, fmt.Errorf("%w: dynamicTable.%s", errUnknownProperty, property)
	}

	return cfg, nil
}

func setStreamProperty(cfg objectConfig, property, value string) (objectConfig, error) {
	switch property {
	case MetadataKeyName:
		cfg.stream.name = value
	case MetadataKeyConsumptionTable:
		cfg.stream.consumptionTable = value
	default:
		return cfg, fmt.Errorf("%w: stream.%s", errUnknownProperty, property)
	}

	return cfg, nil
}
