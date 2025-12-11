package snowflake

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// CDC metadata columns from Snowflake streams.
const (
	metadataAction   = "METADATA$ACTION"
	metadataIsUpdate = "METADATA$ISUPDATE"
	metadataRowID    = "METADATA$ROW_ID"
)

// readMode represents the type of read operation to perform.
type readMode int

const (
	// readModeFullBackfill reads all data from the Dynamic Table.
	// Used when neither Since nor Until is set.
	readModeFullBackfill readMode = iota

	// readModeIncremental reads CDC changes from the Stream.
	// Used when only Since is set (incremental sync).
	// Note: Stream provides changes since last consumption, not since the Since timestamp.
	readModeIncremental

	// readModeTimeRange reads from Dynamic Table with time filtering.
	// Used when both Since and Until are set, or only Until is set.
	readModeTimeRange
)

// determineReadMode determines which read mode to use based on Since/Until parameters.
//
// Read mode logic:
//   - Neither Since nor Until set → Full backfill from Dynamic Table
//   - Only Since set → Incremental read from Stream (CDC)
//   - Only Until set → Time-bounded read from Dynamic Table
//   - Both Since and Until set → Time range read from Dynamic Table
func determineReadMode(params common.ReadParams) readMode {
	hasSince := !params.Since.IsZero()
	hasUntil := !params.Until.IsZero()

	switch {
	case !hasSince && !hasUntil:
		// Full backfill: read everything from Dynamic Table
		return readModeFullBackfill

	case hasSince && !hasUntil:
		// Incremental: read changes from Stream
		return readModeIncremental

	case !hasSince && hasUntil:
		// Historical up to a point: read from Dynamic Table with Until filter
		return readModeTimeRange

	case hasSince && hasUntil:
		// Time range: read from Dynamic Table with Since and Until filters
		return readModeTimeRange

	default:
		// Should never reach here, but default to full backfill
		return readModeFullBackfill
	}
}

// Read reads data from a Snowflake Stream (incremental) or Dynamic Table (full/historical).
//
// Read modes:
//   - Full backfill (no Since/Until): Reads all data from Dynamic Table
//   - Incremental (Since only): Reads CDC changes from Stream
//   - Time range (Until set, or both): Reads from Dynamic Table with time filtering
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	// Get object config from connector's parsed metadata
	objConfig, ok := c.objects.Get(params.ObjectName)
	if !ok {
		return nil, fmt.Errorf("object %q not found in connector configuration", params.ObjectName)
	}

	// Determine read mode based on Since/Until
	mode := determineReadMode(params)

	switch mode {
	case readModeIncremental:
		// Incremental: read from Stream for CDC
		return c.readFromStream(ctx, params, objConfig)

	case readModeFullBackfill, readModeTimeRange:
		// Full backfill or time range: read from Dynamic Table
		return c.readFromDynamicTable(ctx, params, objConfig)

	default:
		// Should never reach here
		return c.readFromDynamicTable(ctx, params, objConfig)
	}
}

// readFromStream reads CDC data from a Snowflake Stream.
func (c *Connector) readFromStream(
	ctx context.Context,
	params common.ReadParams,
	objConfig *objectConfig,
) (*common.ReadResult, error) {
	if objConfig.streamName == "" {
		return nil, fmt.Errorf("streamName not configured for object %q", params.ObjectName)
	}

	streamName := c.getFullyQualifiedName(objConfig.streamName)

	// Build the query to read from stream with CDC metadata
	query := c.buildStreamQuery(streamName, params)

	rows, err := c.handle.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stream: %w", err)
	}
	defer rows.Close()

	// Process rows
	resultRows, err := c.processRows(ctx, rows)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows: int64(len(resultRows)),
		Data: resultRows,
		Done: true, // Streams don't paginate in the traditional sense
	}, nil
}

// readFromDynamicTable reads historical data from a Dynamic Table with time filtering.
func (c *Connector) readFromDynamicTable(
	ctx context.Context,
	params common.ReadParams,
	objConfig *objectConfig,
) (*common.ReadResult, error) {
	if objConfig.dynamicTableName == "" {
		return nil, fmt.Errorf("dynamicTableName not configured for object %q", params.ObjectName)
	}

	tableName := c.getFullyQualifiedName(objConfig.dynamicTableName)

	// Build SELECT query with time filtering
	query := c.buildDynamicTableQuery(tableName, params, objConfig)

	rows, err := c.handle.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query dynamic table: %w", err)
	}
	defer rows.Close()

	// Process rows
	resultRows, err := c.processRows(ctx, rows)
	if err != nil {
		return nil, err
	}

	// Determine if there are more pages
	done := true
	if params.PageSize > 0 && len(resultRows) == params.PageSize {
		// Might be more data - but we don't have cursor-based pagination yet
		done = false
	}

	return &common.ReadResult{
		Rows: int64(len(resultRows)),
		Data: resultRows,
		Done: done,
	}, nil
}

// buildStreamQuery builds the SQL query for reading from a stream.
func (c *Connector) buildStreamQuery(streamName string, params common.ReadParams) string {
	var selectCols string

	if len(params.Fields) > 0 {
		// Select specific fields plus metadata columns
		fields := params.Fields.List()
		quotedFields := make([]string, 0, len(fields)+3)

		for _, f := range fields {
			quotedFields = append(quotedFields, fmt.Sprintf(`"%s"`, strings.ToUpper(f)))
		}

		// Always include CDC metadata columns
		quotedFields = append(quotedFields, metadataAction, metadataIsUpdate, metadataRowID)
		selectCols = strings.Join(quotedFields, ", ")
	} else {
		selectCols = fmt.Sprintf("*, %s, %s, %s", metadataAction, metadataIsUpdate, metadataRowID)
	}

	query := fmt.Sprintf(`SELECT %s FROM %s`, selectCols, streamName)

	// Add LIMIT if PageSize is specified
	if params.PageSize > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, params.PageSize)
	}

	return query
}

// buildDynamicTableQuery builds the SQL query for reading from a dynamic table.
func (c *Connector) buildDynamicTableQuery(
	tableName string,
	params common.ReadParams,
	objConfig *objectConfig,
) string {
	var selectCols string

	if len(params.Fields) > 0 {
		fields := params.Fields.List()
		quotedFields := make([]string, 0, len(fields))

		for _, f := range fields {
			quotedFields = append(quotedFields, fmt.Sprintf(`"%s"`, strings.ToUpper(f)))
		}

		selectCols = strings.Join(quotedFields, ", ")
	} else {
		selectCols = "*"
	}

	query := fmt.Sprintf(`SELECT %s FROM %s`, selectCols, tableName)

	// Add time filtering if timestamp column is specified
	var conditions []string

	if objConfig.timestampColumn != "" {
		if !params.Since.IsZero() {
			conditions = append(conditions,
				fmt.Sprintf(`"%s" >= '%s'`, strings.ToUpper(objConfig.timestampColumn), params.Since.Format("2006-01-02 15:04:05")))
		}

		if !params.Until.IsZero() {
			conditions = append(conditions,
				fmt.Sprintf(`"%s" <= '%s'`, strings.ToUpper(objConfig.timestampColumn), params.Until.Format("2006-01-02 15:04:05")))
		}
	}

	if len(conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}

	// Add ordering by timestamp column if available
	if objConfig.timestampColumn != "" {
		query = fmt.Sprintf("%s ORDER BY \"%s\"", query, strings.ToUpper(objConfig.timestampColumn))
	}

	// Add LIMIT if PageSize is specified
	if params.PageSize > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, params.PageSize)
	}

	return query
}

// processRows converts SQL rows to ReadResultRows.
func (c *Connector) processRows(
	_ context.Context,
	rows *sql.Rows,
) ([]common.ReadResultRow, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var resultRows []common.ReadResultRow

	for rows.Next() {
		row, err := c.scanRow(rows, columns)
		if err != nil {
			return nil, err
		}

		resultRows = append(resultRows, *row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return resultRows, nil
}

// scanRow scans a single row from the result set.
func (c *Connector) scanRow(rows *sql.Rows, columns []string) (*common.ReadResultRow, error) {
	// Create scan destinations
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Build the result row
	fields := make(map[string]any)
	raw := make(map[string]any)

	var rowID string

	for i, col := range columns {
		val := values[i]

		// Convert sql types to Go types
		converted := convertSQLValue(val)

		// Store in raw (always)
		raw[col] = converted

		// Handle CDC metadata columns specially
		switch col {
		case metadataRowID:
			if s, ok := converted.(string); ok {
				rowID = s
			}
			// Also include in fields for downstream processing
			fields[strings.ToLower(col)] = converted
		case metadataAction, metadataIsUpdate:
			fields[strings.ToLower(col)] = converted
		default:
			// For regular columns, store in fields with lowercase key
			fields[strings.ToLower(col)] = converted
		}
	}

	return &common.ReadResultRow{
		Fields: fields,
		Raw:    raw,
		Id:     rowID,
	}, nil
}

// getFullyQualifiedName returns the fully qualified name for an object.
func (c *Connector) getFullyQualifiedName(objectName string) string {
	// If already fully qualified, return as-is
	if strings.Contains(objectName, ".") {
		return objectName
	}

	return fmt.Sprintf(`"%s"."%s"."%s"`,
		strings.ToUpper(c.handle.database),
		strings.ToUpper(c.handle.schema),
		strings.ToUpper(objectName),
	)
}

// convertSQLValue converts SQL types to standard Go types.
func convertSQLValue(val any) any {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []byte:
		// Try to parse as JSON first
		var jsonVal any
		if err := json.Unmarshal(v, &jsonVal); err == nil {
			return jsonVal
		}
		// Otherwise return as string
		return string(v)
	case sql.NullString:
		if v.Valid {
			return v.String
		}

		return nil
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}

		return nil
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64
		}

		return nil
	case sql.NullBool:
		if v.Valid {
			return v.Bool
		}

		return nil
	case sql.NullTime:
		if v.Valid {
			return v.Time
		}

		return nil
	default:
		return v
	}
}
