package snowflake

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type Objects map[string]objectConfig

func (s *Objects) Get(objectName string) (*objectConfig, bool) {
	cfg, ok := (*s)[objectName]
	return &cfg, ok
}

func (s *Objects) ToMetadataMap() map[string]string {
	result := make(map[string]string)

	// Reverse of newSnowflakeObjects
	for objectName, objectConfig := range *s {
		result[fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyQuery)] = objectConfig.query
		result[fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyTimestampColumn)] = objectConfig.timestampColumn
		result[fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyTargetLag)] = objectConfig.targetLag
		result[fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyDynamicTableName)] = objectConfig.dynamicTableName
		result[fmt.Sprintf("%s['%s']['%s']", objectsPrefix, objectName, MetadataKeyStreamName)] = objectConfig.streamName
	}

	return result
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

func newSnowflakeObjects(paramsMap map[string]string) (*Objects, error) {
	result := make(Objects)

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

		// Write the modified config back to the map (Go structs are value types)
		result[objectName] = cfg
	}

	return &result, nil
}

func getStreamName(objectName string) string {
	return fmt.Sprintf("%s%s", objectName, "_stream")
}

func getDynamicTableName(objectName string) string {
	return objectName
}

// EnsureObjects ensures that the objects are created on snowflake.
func (c *Connector) EnsureObjects(ctx context.Context) (*Objects, error) {
	if c.objects == nil {
		return nil, nil
	}

	// Create dynamic tables and streams for each object, and populate their names
	for objectName, objectConfig := range *c.objects {
		// Validate that query exists
		if objectConfig.query == "" {
			return nil, fmt.Errorf("object %q has no query defined", objectName)
		}

		needsUpdate := false

		// Create dynamic table if it doesn't exist
		if objectConfig.dynamicTableName == "" {
			targetLag := objectConfig.targetLag
			if targetLag == "" {
				targetLag = "1 hour"
			}

			dynamicTableName := getDynamicTableName(objectName)
			if err := c.CreateDynamicTable(ctx, dynamicTableName, objectConfig.query, targetLag); err != nil {
				return nil, fmt.Errorf("failed to create dynamic table %s: %w", objectName, err)
			}

			objectConfig.dynamicTableName = dynamicTableName
			needsUpdate = true
		}

		// Create stream if it doesn't exist
		if objectConfig.streamName == "" {
			streamName := getStreamName(objectName)
			if err := c.CreateStream(ctx, streamName, objectName); err != nil {
				return nil, fmt.Errorf("failed to create stream %s: %w", streamName, err)
			}

			objectConfig.streamName = streamName
			needsUpdate = true
		}

		// Only update the map if we made changes
		if needsUpdate {
			(*c.objects)[objectName] = objectConfig
		}
	}

	return c.objects, nil
}
