package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// listObjectMetadata implements the schema lookup.
// This is the internal implementation that gets wired to the DelegateSchemaProvider.
func (c *Connector) listObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
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
	tableName := objectName
	if cfg, ok := c.objects.Get(objectName); ok && cfg.dynamicTableName != "" {
		tableName = cfg.dynamicTableName
	}

	columns, err := c.getColumnMetadata(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get column metadata for %s: %w", objectName, err)
	}

	fields := make(common.FieldsMetadata)
	for _, col := range columns {
		fields[col.Name] = common.FieldMetadata{
			DisplayName:  col.Name,
			ValueType:    snowflakeTypeToValueTypeWithScale(col.DataType, col.NumericScale),
			ProviderType: col.DataType,
			ReadOnly:     boolPtr(false), // Snowflake columns are generally writable if table is
			IsRequired:   boolPtr(!col.IsNullable),
		}
	}

	return common.NewObjectMetadata(objectName, fields), nil
}

// getColumnMetadata retrieves column information for a table/view/dynamic table.
func (c *Connector) getColumnMetadata(ctx context.Context, objectName string) ([]ColumnInfo, error) {
	// Use INFORMATION_SCHEMA.COLUMNS for rich metadata
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

	rows, err := c.handle.db.QueryContext(ctx, query, c.handle.schema, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo

	for rows.Next() {
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
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.IsNullable = isNullable == "YES"

		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}

		if comment.Valid {
			col.Comment = &comment.String
		}

		if charMaxLen.Valid {
			col.CharacterMaxLength = &charMaxLen.Int64
		}

		if numPrecision.Valid {
			col.NumericPrecision = &numPrecision.Int64
		}

		if numScale.Valid {
			col.NumericScale = &numScale.Int64
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("object %s not found or has no columns", objectName)
	}

	return columns, nil
}
