package snowflake

import (
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

// boolPtr returns a pointer to a bool.
func boolPtr(b bool) *bool {
	return &b
}
