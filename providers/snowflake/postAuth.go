package snowflake

import (
	"context"
	"fmt"
)

// Naming constants for Snowflake objects.
// These are used to generate consistent names for Dynamic Tables and Streams.
const (
	// DynamicTableSuffix is appended to object names for Dynamic Tables.
	DynamicTableSuffix = "_DT"

	// StreamSuffix is appended to object names for Streams.
	StreamSuffix = "_STREAM"

	// ObjectNamePrefix is prepended to all Ampersand-managed Snowflake objects.
	ObjectNamePrefix = "AMP_"
)

// DefaultTargetLag is the default target lag for Dynamic Tables.
const DefaultTargetLag = "1 minute"

// Format: AMP_{builderSlug}_{consumerSlug}_{objectName}_DT.
func GenerateDynamicTableName(builderSlug, consumerSlug, objectName string) string {
	return fmt.Sprintf("%s%s_%s_%s%s", ObjectNamePrefix, builderSlug, consumerSlug, objectName, DynamicTableSuffix)
}

// Format: AMP_{builderSlug}_{consumerSlug}_{objectName}_STREAM.
func GenerateStreamName(builderSlug, consumerSlug, objectName string) string {
	return fmt.Sprintf("%s%s_%s_%s%s", ObjectNamePrefix, builderSlug, consumerSlug, objectName, StreamSuffix)
}

// ValidateQuery validates a SQL query by running EXPLAIN.
// This checks syntax and permissions without executing the query.
func (c *Connector) ValidateQuery(ctx context.Context, query string) error {
	explainQuery := "EXPLAIN " + query

	_, err := c.db.ExecContext(ctx, explainQuery)
	if err != nil {
		return fmt.Errorf("query validation failed: %w", err)
	}

	return nil
}

// CreateDynamicTable creates a Dynamic Table from a SQL query.
// The Dynamic Table will automatically refresh based on the target lag.
func (c *Connector) CreateDynamicTable(ctx context.Context, tableName, query, targetLag string) error {
	fqName := c.getFullyQualifiedName(tableName)

	createSQL := fmt.Sprintf(`
		CREATE OR REPLACE DYNAMIC TABLE %s
		TARGET_LAG = '%s'
		WAREHOUSE = %s
		AS %s
	`, fqName, targetLag, c.warehouse, query)

	_, err := c.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create dynamic table %s: %w", fqName, err)
	}

	return nil
}

// DropDynamicTable drops a Dynamic Table if it exists.
func (c *Connector) DropDynamicTable(ctx context.Context, tableName string) error {
	fqName := c.getFullyQualifiedName(tableName)
	dropSQL := "DROP DYNAMIC TABLE IF EXISTS " + fqName

	_, err := c.db.ExecContext(ctx, dropSQL)
	if err != nil {
		return fmt.Errorf("failed to drop dynamic table %s: %w", fqName, err)
	}

	return nil
}

// CreateStream creates a Stream on a Dynamic Table for CDC.
// The stream will include initial rows (SHOW_INITIAL_ROWS = TRUE).
func (c *Connector) CreateStream(ctx context.Context, streamName, dynamicTableName string) error {
	fqStreamName := c.getFullyQualifiedName(streamName)
	fqDTName := c.getFullyQualifiedName(dynamicTableName)

	createSQL := fmt.Sprintf(`
		CREATE OR REPLACE STREAM %s
		ON DYNAMIC TABLE %s
		SHOW_INITIAL_ROWS = TRUE
	`, fqStreamName, fqDTName)

	_, err := c.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create stream %s: %w", fqStreamName, err)
	}

	return nil
}

// DropStream drops a Stream if it exists.
func (c *Connector) DropStream(ctx context.Context, streamName string) error {
	fqName := c.getFullyQualifiedName(streamName)
	dropSQL := "DROP STREAM IF EXISTS " + fqName

	_, err := c.db.ExecContext(ctx, dropSQL)
	if err != nil {
		return fmt.Errorf("failed to drop stream %s: %w", fqName, err)
	}

	return nil
}
