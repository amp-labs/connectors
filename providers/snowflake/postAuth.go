package snowflake

import (
	"context"
	"fmt"
	"strings"
)

// DefaultTargetLag is the default target lag for Dynamic Tables.
const DefaultTargetLag = "1 hour"

// EnsureObjects ensures that the objects are created on snowflake.
func (c *Connector) EnsureObjects(ctx context.Context) (*Objects, error) {
	if c.objects == nil {
		return nil, nil
	}

	// Validate all objects have required configuration
	if err := c.objects.Validate(); err != nil {
		return nil, err
	}

	// Create dynamic tables and streams for each object, and populate their names
	for objectName, cfg := range *c.objects {
		needsUpdate := false

		// Create dynamic table if it doesn't exist
		if cfg.dynamicTable.name == "" {
			// Validate the query before creating the dynamic table
			if err := c.validateQuery(ctx, cfg.query); err != nil {
				return nil, fmt.Errorf("invalid query for object %s: %w", objectName, err)
			}

			targetLag := cfg.dynamicTable.targetLag
			if targetLag == "" {
				targetLag = DefaultTargetLag
			}

			dynamicTableName := getDynamicTableName(objectName)
			if err := c.createDynamicTable(ctx, dynamicTableName, cfg.query, targetLag); err != nil {
				return nil, fmt.Errorf("failed to create dynamic table %s: %w", objectName, err)
			}

			cfg.dynamicTable.name = dynamicTableName
			needsUpdate = true
		}

		// Create stream if it doesn't exist
		if cfg.stream.name == "" {
			streamName := getStreamName(objectName)
			if err := c.createStream(ctx, streamName, objectName); err != nil {
				return nil, fmt.Errorf("failed to create stream %s: %w", streamName, err)
			}

			cfg.stream.name = streamName
			needsUpdate = true
		}

		// Only update the map if we made changes
		if needsUpdate {
			(*c.objects)[objectName] = cfg
		}
	}

	return c.objects, nil
}

// createDynamicTable creates a Dynamic Table from a SQL query.
// The Dynamic Table will automatically refresh based on the target lag.
func (c *Connector) createDynamicTable(ctx context.Context, tableName, query, targetLag string) error {
	fqName := c.getFullyQualifiedName(tableName)

	createSQL := fmt.Sprintf(`
		CREATE OR REPLACE DYNAMIC TABLE %s
		TARGET_LAG = '%s'
		WAREHOUSE = %s
		AS %s
	`, strings.ToUpper(fqName), targetLag, strings.ToUpper(c.handle.warehouse), query)

	_, err := c.handle.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create dynamic table %s: %w", fqName, err)
	}

	return nil
}

// createStream creates a Stream on a Dynamic Table for CDC.
// The stream will include initial rows (SHOW_INITIAL_ROWS = TRUE).
func (c *Connector) createStream(ctx context.Context, streamName, dynamicTableName string) error {
	fqStreamName := c.getFullyQualifiedName(streamName)
	fqDTName := c.getFullyQualifiedName(dynamicTableName)

	createSQL := fmt.Sprintf(`
		CREATE OR REPLACE STREAM %s
		ON DYNAMIC TABLE %s
		SHOW_INITIAL_ROWS = TRUE
	`, strings.ToUpper(fqStreamName), strings.ToUpper(fqDTName))

	_, err := c.handle.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create stream %s: %w", fqStreamName, err)
	}

	return nil
}

// validateQuery validates a SQL query by running EXPLAIN.
// This checks syntax and permissions without executing the query.
func (c *Connector) validateQuery(ctx context.Context, query string) error {
	explainQuery := "EXPLAIN " + query

	rows, err := c.handle.db.QueryContext(ctx, explainQuery)
	if err != nil {
		return fmt.Errorf("query validation failed: %w", err)
	}
	defer rows.Close()

	// Drain the result set to ensure the query completes
	for rows.Next() {
		// We don't need the explain output, just validating
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("query validation failed: %w", err)
	}

	return nil
}
