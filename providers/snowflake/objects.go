package snowflake

import (
	"regexp"
	"strings"
)

type snowflakeObjects map[string]objectConfig

func (s *snowflakeObjects) Get(objectName string) (*objectConfig, bool) {
	cfg, ok := (*s)[objectName]
	return &cfg, ok
}

type objectConfig struct {
	// query is the SQL query defining the data that we treat as an object.
	query string

	// timestampColumn is the column used for incremental filtering on the
	// dynamic table generated over the query.
	timestampColumn string

	// targetLag is the Dynamic Table refresh interval (e.g., "1 hour").
	targetLag string

	// dynamicTableName is the generated DT name (set after PostAuth flow).
	// Used to read data when we need to do non-incremental syncs.
	dynamicTableName string

	// streamName is the generated Stream name (set after PostAuth flow).
	// Used to read data for incremental syncs.
	streamName string
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

const objectsPrefix = "$['objects']"

func newSnowflakeObjects(paramsMap map[string]string) (*snowflakeObjects, error) {
	result := make(snowflakeObjects)

	// Pattern: $['objects']['objectName']['property']
	// Captures: objectName (group 1), property (group 2)
	pattern := regexp.MustCompile(`^\$\['objects'\]\['([^']+)'\]\['([^']+)'\]$`)

	for key, value := range paramsMap {
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
			cfg = objectConfig{}
			result[objectName] = cfg
		}

		// Set the appropriate field based on property name
		switch property {
		case MetadataKeyQuery:
			cfg.query = value
		case MetadataKeyTimestampColumn:
			cfg.timestampColumn = value
		case MetadataKeyTargetLag:
			cfg.targetLag = value
		case MetadataKeyDynamicTableName:
			cfg.dynamicTableName = value
		case MetadataKeyStreamName:
			cfg.streamName = value
		}
	}

	return &result, nil
}
