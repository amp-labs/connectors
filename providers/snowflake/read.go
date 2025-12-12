package snowflake

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// DefaultPageSize is the default number of rows to fetch per page.
const DefaultPageSize = 2000

// CDC metadata columns from Snowflake streams.
const (
	metadataAction   = "METADATA$ACTION"
	metadataIsUpdate = "METADATA$ISUPDATE"
	metadataRowID    = "METADATA$ROW_ID"
)

// Errors for read operations.
var (
	errPrimaryKeyRequired      = fmt.Errorf("primaryKey is required for consistent pagination")
	errTimestampColumnRequired = fmt.Errorf("timestampColumn is required when Since or Until is specified")
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

	// Validate that primaryKey is set (required for consistent pagination)
	if objConfig.primaryKey == "" {
		return nil, errPrimaryKeyRequired
	}

	// Validate that timestampColumn is set if time filtering is requested
	hasTimeFilter := !params.Since.IsZero() || !params.Until.IsZero()
	if hasTimeFilter && objConfig.timestampColumn == "" {
		return nil, errTimestampColumnRequired
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

// AcknowledgeStreamConsumption advances the stream offset by consuming the pending changes.
// This should be called after successfully processing data from readFromStream.
// Note: SELECT alone does not advance the stream offset; only DML operations do.
//
// TODO: Implement the actual consumption logic. This is a stub for now.
func (c *Connector) AcknowledgeStreamConsumption(ctx context.Context, objectName string) error {
	// Get object config
	objConfig, ok := c.objects.Get(objectName)
	if !ok {
		return fmt.Errorf("object %q not found in connector configuration", objectName)
	}

	if objConfig.streamName == "" {
		return fmt.Errorf("streamName not configured for object %q", objectName)
	}

	// TODO: Implement stream consumption via DML operation.
	// Options:
	// 1. MERGE into a staging table
	// 2. INSERT INTO a tracking table (even with no rows)
	// 3. CREATE OR REPLACE STREAM to reset offset
	//
	// For now, this is a stub - the caller must decide how to consume.
	return nil
}

// readFromStream reads CDC data from a Snowflake Stream.
// Note: SELECT from a stream does not advance the stream offset.
// Call AcknowledgeStreamConsumption after processing all pages to advance the offset.
func (c *Connector) readFromStream(
	ctx context.Context,
	params common.ReadParams,
	objConfig *objectConfig,
) (*common.ReadResult, error) {
	if objConfig.streamName == "" {
		return nil, fmt.Errorf("streamName not configured for object %q", params.ObjectName)
	}

	streamName := c.getFullyQualifiedName(objConfig.streamName)

	// Determine page size
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	// Parse offset from NextPage token (default 0)
	offset := 0
	if params.NextPage != "" {
		parsed, err := strconv.Atoi(string(params.NextPage))
		if err != nil {
			return nil, fmt.Errorf("invalid NextPage token: %w", err)
		}

		offset = parsed
	}

	// Build the query to read from stream with CDC metadata
	query := c.buildStreamQuery(streamName, objConfig, pageSize, offset)

	rows, err := c.handle.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stream: %w", err)
	}
	defer rows.Close()

	// Process rows
	resultRows, err := c.processRows(rows, params.Fields)
	if err != nil {
		return nil, err
	}

	// Determine if there might be more data
	done := len(resultRows) < pageSize

	// Calculate next page token
	var nextPage common.NextPageToken
	if !done {
		nextPage = common.NextPageToken(strconv.Itoa(offset + pageSize))
	}

	return &common.ReadResult{
		Rows:     int64(len(resultRows)),
		Data:     resultRows,
		NextPage: nextPage,
		Done:     done,
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

	// Determine page size
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	// Parse offset from NextPage token (default 0)
	offset := 0
	if params.NextPage != "" {
		parsed, err := strconv.Atoi(string(params.NextPage))
		if err != nil {
			return nil, fmt.Errorf("invalid NextPage token: %w", err)
		}

		offset = parsed
	}

	// Build SELECT query with time filtering and pagination
	query := c.buildDynamicTableQuery(tableName, params, objConfig, pageSize, offset)

	rows, err := c.handle.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query dynamic table: %w", err)
	}
	defer rows.Close()

	// Process rows
	resultRows, err := c.processRows(rows, params.Fields)
	if err != nil {
		return nil, err
	}

	// Determine if there might be more data
	done := len(resultRows) < pageSize

	// Calculate next page token
	var nextPage common.NextPageToken
	if !done {
		nextPage = common.NextPageToken(strconv.Itoa(offset + pageSize))
	}

	return &common.ReadResult{
		Rows:     int64(len(resultRows)),
		Data:     resultRows,
		NextPage: nextPage,
		Done:     done,
	}, nil
}

// buildStreamQuery builds the SQL query for reading from a stream.
func (c *Connector) buildStreamQuery(
	streamName string,
	objConfig *objectConfig,
	pageSize int,
	offset int,
) string {
	// Always select all columns plus CDC metadata for streams
	// Fields filtering is done at the result processing level
	query := fmt.Sprintf(`SELECT *, %s, %s, %s FROM %s`,
		metadataAction, metadataIsUpdate, metadataRowID, streamName)

	// Order by primary key for consistent pagination across calls
	query = fmt.Sprintf(`%s ORDER BY "%s" ASC`, query, objConfig.primaryKey)

	// Add LIMIT and OFFSET for pagination
	query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, offset)

	return query
}

// buildDynamicTableQuery builds the SQL query for reading from a dynamic table.
func (c *Connector) buildDynamicTableQuery(
	tableName string,
	params common.ReadParams,
	objConfig *objectConfig,
	pageSize int,
	offset int,
) string {
	// Always select all columns; field filtering is done at result processing level
	query := fmt.Sprintf(`SELECT * FROM %s`, tableName)

	// Build WHERE conditions for time filtering using params.Since and params.Until
	var conditions []string

	if !params.Since.IsZero() {
		conditions = append(conditions,
			fmt.Sprintf(`"%s" >= '%s'`,
				objConfig.timestampColumn,
				params.Since.Format("2006-01-02 15:04:05.999999999")))
	}

	if !params.Until.IsZero() {
		conditions = append(conditions,
			fmt.Sprintf(`"%s" <= '%s'`,
				objConfig.timestampColumn,
				params.Until.Format("2006-01-02 15:04:05.999999999")))
	}

	if len(conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}

	// Order by primary key for consistent pagination across calls
	query = fmt.Sprintf(`%s ORDER BY "%s" ASC`, query, objConfig.primaryKey)

	// Add LIMIT and OFFSET for pagination
	query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, offset)

	return query
}

// processRows converts SQL rows to ReadResultRows.
// requestedFields specifies which fields to include in Fields; if empty, Fields will be empty.
func (c *Connector) processRows(
	rows *sql.Rows,
	requestedFields datautils.StringSet,
) ([]common.ReadResultRow, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Build a set of requested field names for quick lookup
	requestedSet := make(map[string]bool)
	if len(requestedFields) > 0 {
		for _, f := range requestedFields.List() {
			requestedSet[f] = true
		}
	}

	var resultRows []common.ReadResultRow

	for rows.Next() {
		row, err := c.scanRow(rows, columns, requestedSet)
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
// requestedFields is a set of field names to include in Fields; if empty, Fields will be empty.
func (c *Connector) scanRow(rows *sql.Rows, columns []string, requestedFields map[string]bool) (*common.ReadResultRow, error) {
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

		// Store in raw (always includes everything)
		raw[col] = converted

		// Handle CDC metadata columns - extract row ID but don't put in fields
		switch col {
		case metadataRowID:
			if s, ok := converted.(string); ok {
				rowID = s
			}
			// CDC metadata goes in raw only, not fields
		case metadataAction, metadataIsUpdate:
			// CDC metadata goes in raw only, not fields
		default:
			// Only add to fields if it was explicitly requested
			if len(requestedFields) > 0 && requestedFields[col] {
				fields[col] = converted
			}
		}
	}

	return &common.ReadResultRow{
		Fields: fields,
		Raw:    raw,
		Id:     rowID,
	}, nil
}

// getFullyQualifiedName returns the fully qualified name for an object.
// Only the database, schema, and object names are uppercased (FQN components).
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
