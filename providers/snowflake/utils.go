package snowflake

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// ColumnInfo contains metadata about a column.
type ColumnInfo struct {
	// Name is the column name
	Name string

	// DataType is the Snowflake data type (e.g., VARCHAR, NUMBER, TIMESTAMP_NTZ)
	DataType string

	// IsNullable indicates if the column allows NULL values
	IsNullable bool

	// DefaultValue is the column's default value expression, if any
	DefaultValue *string

	// Comment is the column's comment, if any
	Comment *string

	// CharacterMaxLength is the max length for string types
	CharacterMaxLength *int64

	// NumericPrecision is the precision for numeric types
	NumericPrecision *int64

	// NumericScale is the scale for numeric types
	NumericScale *int64
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

func getStreamName(objectName string) string {
	return fmt.Sprintf("%s%s", objectName, "_stream")
}

func getDynamicTableName(objectName string) string {
	return objectName
}

// snowflakeTypeToValueType maps Snowflake data types to Ampersand ValueTypes.
// For NUMBER(p,s): if scale > 0, it's a decimal (Float), otherwise it's an integer (Int).
func snowflakeTypeToValueType(snowflakeType string, scale *int64) common.ValueType {
	// Normalize the type to uppercase for comparison
	upperType := strings.ToUpper(snowflakeType)

	// Handle parameterized types (e.g., VARCHAR(100), NUMBER(10,2))
	if idx := strings.Index(upperType, "("); idx != -1 {
		upperType = upperType[:idx]
	}

	switch upperType {
	// String types
	case "VARCHAR", "TEXT", "STRING", "CHAR", "CHARACTER", "NCHAR", "NVARCHAR", "NVARCHAR2":
		return common.ValueTypeString

	// Integer types - NUMBER without scale defaults to Int
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", "BYTEINT":
		return common.ValueTypeInt

	// NUMBER/NUMERIC/DECIMAL: check scale to determine Int vs Float
	case "NUMBER", "NUMERIC", "DECIMAL":
		if scale != nil && *scale > 0 {
			return common.ValueTypeFloat
		}

		return common.ValueTypeInt

	// Float types
	case "FLOAT", "FLOAT4", "FLOAT8", "DOUBLE", "DOUBLE PRECISION", "REAL", "DECFLOAT":
		return common.ValueTypeFloat

	// Boolean
	case "BOOLEAN":
		return common.ValueTypeBoolean

	// Date only
	case "DATE":
		return common.ValueTypeDate

	// Date/Time types
	case "TIMESTAMP", "TIMESTAMP_LTZ", "TIMESTAMP_NTZ", "TIMESTAMP_TZ",
		"DATETIME", "TIME":
		return common.ValueTypeDateTime

	// Semi-structured types
	case "VARIANT", "OBJECT", "ARRAY", "MAP":
		return common.ValueTypeOther

	// Binary types
	case "BINARY", "VARBINARY":
		return common.ValueTypeOther

	// Geospatial types
	case "GEOGRAPHY", "GEOMETRY":
		return common.ValueTypeOther

	// Vector type (for ML/AI workloads)
	case "VECTOR":
		return common.ValueTypeOther

	// File type (unstructured data reference)
	case "FILE":
		return common.ValueTypeOther

	default:
		return common.ValueTypeOther
	}
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

// boolPtr returns a pointer to a bool.
func boolPtr(b bool) *bool {
	return &b
}
