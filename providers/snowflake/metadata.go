package snowflake

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var errObjectsNotInitialized = errors.New("object(s) not initialized")

// listObjectMetadata implements the schema lookup.
// This is the internal implementation that gets wired to the DelegateSchemaProvider.
func (c *Connector) listObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if c.objects == nil {
		return nil, errObjectsNotInitialized
	}

	result := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		metadata, err := c.getObjectMetadata(ctx, objectName)
		if err != nil {
			result.AppendError(objectName, err)

			continue
		}

		result.Result[objectName] = *metadata
	}

	return result, nil
}

// getObjectMetadata retrieves metadata for a single object.
func (c *Connector) getObjectMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	// Resolve the actual table name from object config if available
	cfg, ok := c.objects.Get(objectName)
	if !ok || cfg.dynamicTable.name == "" {
		return nil, fmt.Errorf("%w: %q", errObjectsNotInitialized, objectName)
	}

	columns, err := c.getColumnMetadata(ctx, cfg.dynamicTable.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get column metadata for %s: %w", cfg.dynamicTable.name, err)
	}

	fields := make(common.FieldsMetadata)

	for _, col := range columns {
		// Lowercase field names for consistency.
		// Snowflake returns UPPERCASE column names, but we normalize to lowercase
		// so that field selection, mappings, and webhook delivery are all consistent.
		fieldName := strings.ToLower(col.Name)
		fields[fieldName] = common.FieldMetadata{
			DisplayName:  col.Name, // Keep original casing for display
			ValueType:    snowflakeTypeToValueType(col.DataType, col.NumericScale),
			ProviderType: col.DataType,
			ReadOnly:     boolPtr(false), // Snowflake columns are generally writable if table is
			IsRequired:   boolPtr(!col.IsNullable),
		}
	}

	return common.NewObjectMetadata(objectName, fields), nil
}

// getColumnMetadata retrieves column information for a table/view/dynamic table.
func (c *Connector) getColumnMetadata(ctx context.Context, objectName string) ([]ColumnInfo, error) {
	rows, err := c.queryColumnMetadata(ctx, objectName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := c.scanColumnRows(rows)
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("%w: %s", errObjectNoColumns, objectName)
	}

	return columns, nil
}

func (c *Connector) queryColumnMetadata(ctx context.Context, objectName string) (*sql.Rows, error) {
	// Use INFORMATION_SCHEMA.COLUMNS for rich metadata.
	// The database name comes from our validated connection configuration, not user input.
	//nolint:gosec // database name is from validated config, not user input
	query := fmt.Sprintf(`
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COMMENT,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE
		FROM %s.INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, c.handle.database)

	rows, err := c.handle.db.QueryContext(ctx, query, strings.ToUpper(c.handle.schema), strings.ToUpper(objectName))
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}

	return rows, nil
}

func (c *Connector) scanColumnRows(rows *sql.Rows) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	for rows.Next() {
		col, err := c.scanSingleColumn(rows)
		if err != nil {
			return nil, err
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return columns, nil
}

func (c *Connector) scanSingleColumn(rows *sql.Rows) (ColumnInfo, error) {
	var col ColumnInfo

	var isNullable string

	var defaultValue, comment sql.NullString

	var charMaxLen, numPrecision, numScale sql.NullInt64

	err := rows.Scan(
		&col.Name,
		&col.DataType,
		&isNullable,
		&defaultValue,
		&comment,
		&charMaxLen,
		&numPrecision,
		&numScale,
	)
	if err != nil {
		return ColumnInfo{}, fmt.Errorf("failed to scan column: %w", err)
	}

	col.IsNullable = isNullable == "YES"
	col.DefaultValue = nullStringToPtr(defaultValue)
	col.Comment = nullStringToPtr(comment)
	col.CharacterMaxLength = nullInt64ToPtr(charMaxLen)
	col.NumericPrecision = nullInt64ToPtr(numPrecision)
	col.NumericScale = nullInt64ToPtr(numScale)

	return col, nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}

	return nil
}

func nullInt64ToPtr(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}

	return nil
}
