package snowflake

import (
	"context"
	"fmt"
	"strings"
)

// DefaultTargetLag is the default target lag for Dynamic Tables.
const DefaultTargetLag = "1 hour"

// StreamConsumptionTable is the name of the table used to advance stream offsets.
// This table is created automatically during EnsureObjects and is used by
// AcknowledgeStreamConsumption to advance stream offsets without storing data.
const StreamConsumptionTable = "_AMP_STREAM_CONSUMPTION"

// EnsureObjects ensures that the objects are created on snowflake.
// Returns the updated Objects or nil if no objects are configured.
func (c *Connector) EnsureObjects(ctx context.Context) (*Objects, error) {
	if c.objects == nil {
		return nil, nil //nolint:nilnil // nil objects is a valid state (no objects configured)
	}

	// Validate all objects have required configuration
	if err := c.objects.Validate(); err != nil {
		return nil, err
	}

	// Create the stream consumption table (used by AcknowledgeStreamConsumption).
	// This is a single table shared by all streams.
	if err := c.ensureStreamConsumptionTable(ctx); err != nil {
		return nil, err
	}

	// Create dynamic tables and streams for each object, and populate their names
	for objectName, cfg := range *c.objects {
		updatedCfg, err := c.ensureSingleObject(ctx, objectName, cfg)
		if err != nil {
			return nil, err
		}

		if updatedCfg != nil {
			(*c.objects)[objectName] = *updatedCfg
		}
	}

	return c.objects, nil
}

// ensureStreamConsumptionTable creates the stream consumption table if it doesn't exist.
// This table is used by AcknowledgeStreamConsumption to advance stream offsets
// without actually storing any data (using INSERT ... WHERE FALSE).
func (c *Connector) ensureStreamConsumptionTable(ctx context.Context) error {
	fqName := c.getFullyQualifiedName(StreamConsumptionTable)

	// Create a minimal table with a single nullable column.
	// We never actually insert data - the INSERT ... WHERE FALSE pattern
	// advances the stream offset without writing anything.
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			_placeholder NUMBER
		)
	`, strings.ToUpper(fqName))

	_, err := c.handle.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create stream consumption table: %w", err)
	}

	return nil
}

func (c *Connector) ensureSingleObject(
	ctx context.Context, objectName string, cfg objectConfig,
) (*objectConfig, error) {
	needsUpdate := false

	// Create dynamic table if it doesn't exist
	if cfg.dynamicTable.name == "" {
		if err := c.ensureDynamicTable(ctx, objectName, &cfg); err != nil {
			return nil, err
		}

		needsUpdate = true
	}

	// Create stream if it doesn't exist
	if cfg.stream.name == "" {
		if err := c.ensureStream(ctx, objectName, &cfg); err != nil {
			return nil, err
		}

		needsUpdate = true
	}

	if needsUpdate {
		return &cfg, nil
	}

	return nil, nil //nolint:nilnil // nil indicates no update needed
}

func (c *Connector) ensureDynamicTable(ctx context.Context, objectName string, cfg *objectConfig) error {
	// Validate the query before creating the dynamic table
	if err := c.validateQuery(ctx, cfg.dynamicTable.query); err != nil {
		return fmt.Errorf("invalid query for object %s: %w", objectName, err)
	}

	targetLag := cfg.dynamicTable.targetLag
	if targetLag == "" {
		targetLag = DefaultTargetLag
	}

	dynamicTableName := getDynamicTableName(objectName)
	if err := c.createDynamicTable(ctx, dynamicTableName, cfg.dynamicTable.query, targetLag); err != nil {
		return fmt.Errorf("failed to create dynamic table %s: %w", objectName, err)
	}

	cfg.dynamicTable.name = dynamicTableName

	return nil
}

func (c *Connector) ensureStream(ctx context.Context, objectName string, cfg *objectConfig) error {
	streamName := getStreamName(objectName)
	if err := c.createStream(ctx, streamName, objectName); err != nil {
		return fmt.Errorf("failed to create stream %s: %w", streamName, err)
	}

	cfg.stream.name = streamName
	cfg.stream.consumptionTable = StreamConsumptionTable

	return nil
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
// SHOW_INITIAL_ROWS = FALSE means the stream only captures changes AFTER creation.
// For initial data, users should do a backfill read from the Dynamic Table first.
// This avoids duplicate data when users do backfill followed by incremental reads.
//
// IMPORTANT: Streams become stale if not consumed within the data retention period
// (default 1 day, up to 90 days for Enterprise). Schedule syncs accordingly.
func (c *Connector) createStream(ctx context.Context, streamName, dynamicTableName string) error {
	fqStreamName := c.getFullyQualifiedName(streamName)
	fqDTName := c.getFullyQualifiedName(dynamicTableName)

	createSQL := fmt.Sprintf(`
		CREATE OR REPLACE STREAM %s
		ON DYNAMIC TABLE %s
		SHOW_INITIAL_ROWS = FALSE
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
