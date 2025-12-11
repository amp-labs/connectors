package snowflake

import (
	"context"
	"fmt"
	"strings"
)

// DefaultTargetLag is the default target lag for Dynamic Tables.
const DefaultTargetLag = "1 minute"

// ValidateQuery validates a SQL query by running EXPLAIN.
// This checks syntax and permissions without executing the query.
func (c *Connector) ValidateQuery(ctx context.Context, query string) error {
	explainQuery := "EXPLAIN " + query

	_, err := c.handle.db.ExecContext(ctx, explainQuery)
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
	`, strings.ToUpper(fqName), targetLag, strings.ToUpper(c.handle.warehouse), query)

	_, err := c.handle.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create dynamic table %s: %w", fqName, err)
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
	`, strings.ToUpper(fqStreamName), strings.ToUpper(fqDTName))

	_, err := c.handle.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create stream %s: %w", fqStreamName, err)
	}

	return nil
}
