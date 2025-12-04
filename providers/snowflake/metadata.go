package snowflake

import (
	"regexp"
	"strings"
)

// ObjectConfig holds per-object configuration parsed from metadata.
// These values come from JSONPath-style keys in providerMetadata.
type ObjectConfig struct {
	// Query is the SQL query defining the data (set by builder).
	Query string

	// TimestampColumn is the column used for incremental filtering.
	TimestampColumn string

	// TargetLag is the Dynamic Table refresh interval (e.g., "1 hour").
	TargetLag string

	// DynamicTableName is the generated DT name (set after PostAuth).
	DynamicTableName string

	// StreamName is the generated Stream name (set after PostAuth).
	StreamName string
}

// Metadata keys for per-object configuration.
// These are the property names expected in JSONPath keys like $['objects']['objName']['query'].
const (
	MetadataKeyQuery            = "query"
	MetadataKeyTimestampColumn  = "timestampColumn"
	MetadataKeyTargetLag        = "targetLag"
	MetadataKeyDynamicTableName = "dynamicTableName"
	MetadataKeyStreamName       = "streamName"
)

// objectsPrefix is the JSONPath prefix for per-object configuration.
const objectsPrefix = "$['objects']"

// parseObjectConfigs extracts per-object configurations from flat metadata keys.
// It looks for keys matching the pattern $['objects']['objectName']['property']
// and builds a map of objectName -> ObjectConfig.
//
// Example input:
//
//	map[string]string{
//	  "warehouse": "WH1",
//	  "$['objects']['contacts']['query']": "SELECT * FROM ...",
//	  "$['objects']['contacts']['timestampColumn']": "updated_at",
//	  "$['objects']['accounts']['query']": "SELECT * FROM ...",
//	}
//
// Example output:
//
//	map[string]*ObjectConfig{
//	  "contacts": {Query: "SELECT * FROM ...", TimestampColumn: "updated_at"},
//	  "accounts": {Query: "SELECT * FROM ..."},
//	}
func parseObjectConfigs(metadata map[string]string) map[string]*ObjectConfig {
	result := make(map[string]*ObjectConfig)

	// Pattern: $['objects']['objectName']['property']
	// Captures: objectName (group 1), property (group 2)
	pattern := regexp.MustCompile(`^\$\['objects'\]\['([^']+)'\]\['([^']+)'\]$`)

	for key, value := range metadata {
		// Skip keys that don't start with objects prefix
		if !strings.HasPrefix(key, objectsPrefix) {
			continue
		}

		matches := pattern.FindStringSubmatch(key)
		if len(matches) != 3 {
			continue
		}

		objectName := matches[1]
		property := matches[2]

		// Get or create ObjectConfig for this object
		cfg, ok := result[objectName]
		if !ok {
			cfg = &ObjectConfig{}
			result[objectName] = cfg
		}

		// Set the appropriate field based on property name
		switch property {
		case MetadataKeyQuery:
			cfg.Query = value
		case MetadataKeyTimestampColumn:
			cfg.TimestampColumn = value
		case MetadataKeyTargetLag:
			cfg.TargetLag = value
		case MetadataKeyDynamicTableName:
			cfg.DynamicTableName = value
		case MetadataKeyStreamName:
			cfg.StreamName = value
		}
	}

	return result
}
