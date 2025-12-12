package snowflake

import (
	"fmt"
	"regexp"
	"strings"
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
		if cfg.query == "" {
			errors = append(errors, fmt.Sprintf("object %q: missing required field '%s'",
				objectName, MetadataKeyQuery))
		}

		if cfg.dynamicTable.primaryKey == "" {
			errors = append(errors, fmt.Sprintf("object %q: missing required field '%s.%s'",
				objectName, MetadataKeyDynamicTable, MetadataKeyPrimaryKey))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("snowflake objects validation failed:\n  - %s", strings.Join(errors, "\n  - "))
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
	// Structure: $['objects']['objName']['query'] for object-level
	//            $['objects']['objName']['dynamicTable']['property'] for DT
	//            $['objects']['objName']['stream']['property'] for Stream
	for objectName, cfg := range *s {
		// Object-level property
		addIfNotEmpty(fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyQuery), cfg.query)

		// Dynamic Table properties (nested under 'dynamicTable')
		dtPrefix := fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyDynamicTable)
		addIfNotEmpty(fmt.Sprintf("%s['%s']", dtPrefix, MetadataKeyPrimaryKey), cfg.dynamicTable.primaryKey)
		addIfNotEmpty(fmt.Sprintf("%s['%s']", dtPrefix, MetadataKeyTimestampColumn), cfg.dynamicTable.timestampColumn)
		addIfNotEmpty(fmt.Sprintf("%s['%s']", dtPrefix, MetadataKeyTargetLag), cfg.dynamicTable.targetLag)
		addIfNotEmpty(fmt.Sprintf("%s['%s']", dtPrefix, MetadataKeyName), cfg.dynamicTable.name)

		// Stream properties (nested under 'stream')
		streamPrefix := fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyStream)
		addIfNotEmpty(fmt.Sprintf("%s['%s']", streamPrefix, MetadataKeyName), cfg.stream.name)
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
//	$['objects']['objName']['query']
//	$['objects']['objName']['dynamicTable'][...]
//	$['objects']['objName']['stream'][...]
type objectConfig struct {
	// query is the SQL query defining the data that we treat as an object.
	query string

	// dynamicTable holds Dynamic Table specific configuration.
	dynamicTable dynamicTableConfig

	// stream holds Stream specific configuration.
	stream streamConfig
}

// dynamicTableConfig holds configuration for a Snowflake Dynamic Table.
type dynamicTableConfig struct {
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
}

// Metadata keys for per-object configuration.
// These are the property names expected in JSONPath keys.
//
// Structure:
//
//	$['objects']['objName']['query']                      - SQL query (object level)
//	$['objects']['objName']['dynamicTable']['primaryKey'] - Primary key column
//	$['objects']['objName']['dynamicTable']['timestampColumn'] - Timestamp column
//	$['objects']['objName']['dynamicTable']['targetLag'] - Refresh interval
//	$['objects']['objName']['dynamicTable']['name']      - Generated DT name
//	$['objects']['objName']['stream']['name']            - Generated stream name
const (
	// MetadataKeyQuery is the SQL query defining the data source (object level).
	MetadataKeyQuery = "query"

	// Parent keys for nested structure.
	MetadataKeyDynamicTable = "dynamicTable"
	MetadataKeyStream       = "stream"

	// Dynamic Table specific keys (nested under 'dynamicTable').
	MetadataKeyPrimaryKey      = "primaryKey"
	MetadataKeyTimestampColumn = "timestampColumn"
	MetadataKeyTargetLag       = "targetLag"
	MetadataKeyName            = "name"
)

const objectsPrefix = "$['objects']"

func newSnowflakeObjects(paramsMap map[string]string) (*Objects, error) {
	result := make(Objects)

	// Pattern for 3-level paths: $['objects']['objectName']['property']
	// Used for object-level properties like 'query'
	// Captures: objectName (group 1), property (group 2)
	pattern3Level := regexp.MustCompile(`^\$\['objects'\]\['([^']+)'\]\['([^']+)'\]$`)

	// Pattern for 4-level paths: $['objects']['objectName']['parent']['property']
	// Used for nested properties like 'dynamicTable.primaryKey' or 'stream.name'
	// Captures: objectName (group 1), parent (group 2), property (group 3)
	pattern4Level := regexp.MustCompile(`^\$\['objects'\]\['([^']+)'\]\['([^']+)'\]\['([^']+)'\]$`)

	for key, value := range paramsMap {
		// Skip keys that don't start with objects prefix
		if !strings.HasPrefix(key, objectsPrefix) {
			continue
		}

		// Try 4-level pattern first (more specific)
		if matches := pattern4Level.FindStringSubmatch(key); len(matches) == 4 {
			objectName := matches[1]
			parent := matches[2]
			property := matches[3]

			cfg := result[objectName]

			switch parent {
			case MetadataKeyDynamicTable:
				switch property {
				case MetadataKeyPrimaryKey:
					cfg.dynamicTable.primaryKey = value
				case MetadataKeyTimestampColumn:
					cfg.dynamicTable.timestampColumn = value
				case MetadataKeyTargetLag:
					cfg.dynamicTable.targetLag = value
				case MetadataKeyName:
					cfg.dynamicTable.name = value
				}
			case MetadataKeyStream:
				switch property {
				case MetadataKeyName:
					cfg.stream.name = value
				}
			}

			result[objectName] = cfg

			continue
		}

		// Try 3-level pattern (object-level properties)
		if matches := pattern3Level.FindStringSubmatch(key); len(matches) == 3 {
			objectName := matches[1]
			property := matches[2]

			cfg := result[objectName]

			switch property {
			case MetadataKeyQuery:
				cfg.query = value
			}

			result[objectName] = cfg
		}
	}

	return &result, nil
}
