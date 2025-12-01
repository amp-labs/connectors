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
// Note: For NUMBER types with scale (e.g., NUMBER(10,2)), use snowflakeTypeToValueTypeWithScale
// to correctly identify decimal types.
func snowflakeTypeToValueType(snowflakeType string) common.ValueType {
	// Normalize the type to uppercase for comparison
	upperType := strings.ToUpper(snowflakeType)

	// Handle parameterized types (e.g., VARCHAR(100), NUMBER(10,2))
	if idx := strings.Index(upperType, "("); idx != -1 {
		upperType = upperType[:idx]
	}

	switch upperType {
	case "VARCHAR", "TEXT", "STRING", "CHAR", "CHARACTER":
		return common.ValueTypeString

	// NUMBER without scale info defaults to Int - caller should use
	// snowflakeTypeToValueTypeWithScale for accurate NUMBER handling
	case "NUMBER", "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", "BYTEINT":
		return common.ValueTypeInt

	case "FLOAT", "FLOAT4", "FLOAT8", "DOUBLE", "DOUBLE PRECISION", "REAL":
		return common.ValueTypeFloat

	case "BOOLEAN":
		return common.ValueTypeBoolean

	case "DATE":
		return common.ValueTypeDate

	case "TIMESTAMP", "TIMESTAMP_LTZ", "TIMESTAMP_NTZ", "TIMESTAMP_TZ",
		"DATETIME", "TIME":
		return common.ValueTypeDateTime

	case "VARIANT", "OBJECT", "ARRAY":
		return common.ValueTypeOther

	case "BINARY", "VARBINARY":
		return common.ValueTypeOther

	case "GEOGRAPHY", "GEOMETRY":
		return common.ValueTypeOther

	default:
		return common.ValueTypeOther
	}
}

// snowflakeTypeToValueTypeWithScale maps Snowflake data types to Ampersand ValueTypes,
// correctly handling NUMBER types with scale information.
// For NUMBER(p,s): if scale > 0, it's a decimal (Float), otherwise it's an integer (Int).
func snowflakeTypeToValueTypeWithScale(snowflakeType string, scale *int64) common.ValueType {
	upperType := strings.ToUpper(snowflakeType)

	// Handle parameterized types
	if idx := strings.Index(upperType, "("); idx != -1 {
		upperType = upperType[:idx]
	}

	// Special handling for NUMBER with scale
	if upperType == "NUMBER" && scale != nil && *scale > 0 {
		return common.ValueTypeFloat
	}

	return snowflakeTypeToValueType(snowflakeType)
}

// boolPtr returns a pointer to a bool.
func boolPtr(b bool) *bool {
	return &b
}
