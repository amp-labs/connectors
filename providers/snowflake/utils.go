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
	return objectName + "_stream"
}

func getDynamicTableName(objectName string) string {
	return objectName
}

// Type mapping tables for Snowflake to Ampersand value types.
var (
	stringTypes   = map[string]bool{"VARCHAR": true, "TEXT": true, "STRING": true, "CHAR": true, "CHARACTER": true, "NCHAR": true, "NVARCHAR": true, "NVARCHAR2": true}          //nolint:gochecknoglobals
	integerTypes  = map[string]bool{"INT": true, "INTEGER": true, "BIGINT": true, "SMALLINT": true, "TINYINT": true, "BYTEINT": true}                                            //nolint:gochecknoglobals
	numericTypes  = map[string]bool{"NUMBER": true, "NUMERIC": true, "DECIMAL": true}                                                                                            //nolint:gochecknoglobals
	floatTypes    = map[string]bool{"FLOAT": true, "FLOAT4": true, "FLOAT8": true, "DOUBLE": true, "DOUBLE PRECISION": true, "REAL": true, "DECFLOAT": true}                     //nolint:gochecknoglobals
	dateTimeTypes = map[string]bool{"TIMESTAMP": true, "TIMESTAMP_LTZ": true, "TIMESTAMP_NTZ": true, "TIMESTAMP_TZ": true, "DATETIME": true, "TIME": true}                       //nolint:gochecknoglobals
	otherTypes    = map[string]bool{"VARIANT": true, "OBJECT": true, "ARRAY": true, "MAP": true, "BINARY": true, "VARBINARY": true, "GEOGRAPHY": true, "GEOMETRY": true, "VECTOR": true, "FILE": true} //nolint:gochecknoglobals
)

// snowflakeTypeToValueType maps Snowflake data types to Ampersand ValueTypes.
// For NUMBER(p,s): if scale > 0, it's a decimal (Float), otherwise it's an integer (Int).
func snowflakeTypeToValueType(snowflakeType string, scale *int64) common.ValueType {
	// Normalize the type to uppercase for comparison
	upperType := strings.ToUpper(snowflakeType)

	// Handle parameterized types (e.g., VARCHAR(100), NUMBER(10,2))
	if idx := strings.Index(upperType, "("); idx != -1 {
		upperType = upperType[:idx]
	}

	return mapSnowflakeType(upperType, scale)
}

func mapSnowflakeType(upperType string, scale *int64) common.ValueType {
	switch {
	case stringTypes[upperType]:
		return common.ValueTypeString
	case integerTypes[upperType]:
		return common.ValueTypeInt
	case numericTypes[upperType]:
		return mapNumericType(scale)
	case floatTypes[upperType]:
		return common.ValueTypeFloat
	case upperType == "BOOLEAN":
		return common.ValueTypeBoolean
	case upperType == "DATE":
		return common.ValueTypeDate
	case dateTimeTypes[upperType]:
		return common.ValueTypeDateTime
	case otherTypes[upperType]:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

func mapNumericType(scale *int64) common.ValueType {
	if scale != nil && *scale > 0 {
		return common.ValueTypeFloat
	}

	return common.ValueTypeInt
}

// convertSQLValue converts SQL types to standard Go types.
func convertSQLValue(value any) any {
	if value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []byte:
		return convertBytes(typed)
	case sql.NullString:
		return convertNullString(typed)
	case sql.NullInt64:
		return convertNullInt64(typed)
	case sql.NullFloat64:
		return convertNullFloat64(typed)
	case sql.NullBool:
		return convertNullBool(typed)
	case sql.NullTime:
		return convertNullTime(typed)
	default:
		return typed
	}
}

func convertBytes(data []byte) any {
	// Try to parse as JSON first
	var jsonVal any
	if err := json.Unmarshal(data, &jsonVal); err == nil {
		return jsonVal
	}
	// Otherwise return as string
	return string(data)
}

func convertNullString(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}

	return nil
}

func convertNullInt64(ni sql.NullInt64) any {
	if ni.Valid {
		return ni.Int64
	}

	return nil
}

func convertNullFloat64(nf sql.NullFloat64) any {
	if nf.Valid {
		return nf.Float64
	}

	return nil
}

func convertNullBool(nb sql.NullBool) any {
	if nb.Valid {
		return nb.Bool
	}

	return nil
}

func convertNullTime(nt sql.NullTime) any {
	if nt.Valid {
		return nt.Time
	}

	return nil
}

// boolPtr returns a pointer to a bool.
func boolPtr(b bool) *bool {
	return &b
}
