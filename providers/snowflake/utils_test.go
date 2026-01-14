package snowflake

import (
	"database/sql"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
)

func TestGetFullyQualifiedName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		database   string
		schema     string
		objectName string
		want       string
	}{
		{
			name:       "simple object name",
			database:   "mydb",
			schema:     "myschema",
			objectName: "users",
			want:       `"MYDB"."MYSCHEMA"."USERS"`,
		},
		{
			name:       "lowercase inputs get uppercased",
			database:   "testdb",
			schema:     "public",
			objectName: "customers",
			want:       `"TESTDB"."PUBLIC"."CUSTOMERS"`,
		},
		{
			name:       "mixed case inputs",
			database:   "MyDatabase",
			schema:     "MySchema",
			objectName: "MyTable",
			want:       `"MYDATABASE"."MYSCHEMA"."MYTABLE"`,
		},
		{
			name:       "already fully qualified returns as-is",
			database:   "db",
			schema:     "schema",
			objectName: "other.schema.table",
			want:       "other.schema.table",
		},
		{
			name:       "object with single dot returns as-is",
			database:   "db",
			schema:     "schema",
			objectName: "schema.table",
			want:       "schema.table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Connector{
				handle: &connectionInfo{
					database: tt.database,
					schema:   tt.schema,
				},
			}

			got := c.getFullyQualifiedName(tt.objectName)
			if got != tt.want {
				t.Errorf("getFullyQualifiedName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetStreamName(t *testing.T) {
	t.Parallel()

	name, err := getStreamName("contacts")
	if err != nil {
		t.Fatalf("getStreamName() error = %v", err)
	}

	// Should have format: objectName_stream_suffix
	if len(name) < len("contacts_stream_") {
		t.Errorf("getStreamName() name too short: %q", name)
	}

	// Check prefix
	expectedPrefix := "contacts_stream_"
	if name[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("getStreamName() = %q, want prefix %q", name, expectedPrefix)
	}

	// Suffix should be 6 hex characters
	suffix := name[len(expectedPrefix):]
	if len(suffix) != 6 {
		t.Errorf("getStreamName() suffix length = %d, want 6", len(suffix))
	}
}

func TestGetDynamicTableName(t *testing.T) {
	t.Parallel()

	name, err := getDynamicTableName("orders")
	if err != nil {
		t.Fatalf("getDynamicTableName() error = %v", err)
	}

	// Should have format: objectName_dt_suffix
	expectedPrefix := "orders_dt_"
	if len(name) < len(expectedPrefix) {
		t.Errorf("getDynamicTableName() name too short: %q", name)
	}

	if name[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("getDynamicTableName() = %q, want prefix %q", name, expectedPrefix)
	}

	// Suffix should be 6 hex characters
	suffix := name[len(expectedPrefix):]
	if len(suffix) != 6 {
		t.Errorf("getDynamicTableName() suffix length = %d, want 6", len(suffix))
	}
}

func TestGenerateRandomSuffix(t *testing.T) {
	t.Parallel()

	// Generate multiple suffixes and ensure they're unique
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		suffix, err := generateRandomSuffix()
		if err != nil {
			t.Fatalf("generateRandomSuffix() error = %v", err)
		}

		// Should be 6 hex characters (3 bytes = 6 hex chars)
		if len(suffix) != 6 {
			t.Errorf("generateRandomSuffix() length = %d, want 6", len(suffix))
		}

		// Should be unique
		if seen[suffix] {
			t.Errorf("generateRandomSuffix() produced duplicate: %q", suffix)
		}

		seen[suffix] = true
	}
}

func TestSnowflakeTypeToValueType(t *testing.T) {
	t.Parallel()

	// Helper to create int64 pointer
	int64Ptr := func(v int64) *int64 { return &v }

	tests := []struct {
		name          string
		snowflakeType string
		scale         *int64
		want          common.ValueType
	}{
		// String types
		{name: "VARCHAR", snowflakeType: "VARCHAR", want: common.ValueTypeString},
		{name: "VARCHAR with size", snowflakeType: "VARCHAR(255)", want: common.ValueTypeString},
		{name: "TEXT", snowflakeType: "TEXT", want: common.ValueTypeString},
		{name: "STRING", snowflakeType: "STRING", want: common.ValueTypeString},
		{name: "CHAR", snowflakeType: "CHAR", want: common.ValueTypeString},
		{name: "NVARCHAR", snowflakeType: "NVARCHAR", want: common.ValueTypeString},

		// Integer types
		{name: "INT", snowflakeType: "INT", want: common.ValueTypeInt},
		{name: "INTEGER", snowflakeType: "INTEGER", want: common.ValueTypeInt},
		{name: "BIGINT", snowflakeType: "BIGINT", want: common.ValueTypeInt},
		{name: "SMALLINT", snowflakeType: "SMALLINT", want: common.ValueTypeInt},
		{name: "TINYINT", snowflakeType: "TINYINT", want: common.ValueTypeInt},

		// Numeric types (depends on scale)
		{name: "NUMBER no scale", snowflakeType: "NUMBER", scale: nil, want: common.ValueTypeInt},
		{name: "NUMBER scale 0", snowflakeType: "NUMBER", scale: int64Ptr(0), want: common.ValueTypeInt},
		{name: "NUMBER with scale", snowflakeType: "NUMBER", scale: int64Ptr(2), want: common.ValueTypeFloat},
		{name: "DECIMAL no scale", snowflakeType: "DECIMAL", scale: nil, want: common.ValueTypeInt},
		{name: "DECIMAL with scale", snowflakeType: "DECIMAL", scale: int64Ptr(4), want: common.ValueTypeFloat},
		{name: "NUMBER(10,2) parameterized", snowflakeType: "NUMBER(10,2)", scale: int64Ptr(2), want: common.ValueTypeFloat},

		// Float types
		{name: "FLOAT", snowflakeType: "FLOAT", want: common.ValueTypeFloat},
		{name: "FLOAT4", snowflakeType: "FLOAT4", want: common.ValueTypeFloat},
		{name: "FLOAT8", snowflakeType: "FLOAT8", want: common.ValueTypeFloat},
		{name: "DOUBLE", snowflakeType: "DOUBLE", want: common.ValueTypeFloat},
		{name: "DOUBLE PRECISION", snowflakeType: "DOUBLE PRECISION", want: common.ValueTypeFloat},
		{name: "REAL", snowflakeType: "REAL", want: common.ValueTypeFloat},

		// Boolean
		{name: "BOOLEAN", snowflakeType: "BOOLEAN", want: common.ValueTypeBoolean},
		{name: "boolean lowercase", snowflakeType: "boolean", want: common.ValueTypeBoolean},

		// Date types
		{name: "DATE", snowflakeType: "DATE", want: common.ValueTypeDate},

		// DateTime types
		{name: "TIMESTAMP", snowflakeType: "TIMESTAMP", want: common.ValueTypeDateTime},
		{name: "TIMESTAMP_LTZ", snowflakeType: "TIMESTAMP_LTZ", want: common.ValueTypeDateTime},
		{name: "TIMESTAMP_NTZ", snowflakeType: "TIMESTAMP_NTZ", want: common.ValueTypeDateTime},
		{name: "TIMESTAMP_TZ", snowflakeType: "TIMESTAMP_TZ", want: common.ValueTypeDateTime},
		{name: "DATETIME", snowflakeType: "DATETIME", want: common.ValueTypeDateTime},
		{name: "TIME", snowflakeType: "TIME", want: common.ValueTypeDateTime},

		// Other/complex types
		{name: "VARIANT", snowflakeType: "VARIANT", want: common.ValueTypeOther},
		{name: "OBJECT", snowflakeType: "OBJECT", want: common.ValueTypeOther},
		{name: "ARRAY", snowflakeType: "ARRAY", want: common.ValueTypeOther},
		{name: "BINARY", snowflakeType: "BINARY", want: common.ValueTypeOther},
		{name: "GEOGRAPHY", snowflakeType: "GEOGRAPHY", want: common.ValueTypeOther},
		{name: "GEOMETRY", snowflakeType: "GEOMETRY", want: common.ValueTypeOther},

		// Unknown type defaults to Other
		{name: "UNKNOWN_TYPE", snowflakeType: "UNKNOWN_TYPE", want: common.ValueTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := snowflakeTypeToValueType(tt.snowflakeType, tt.scale)
			if got != tt.want {
				t.Errorf("snowflakeTypeToValueType(%q, %v) = %v, want %v",
					tt.snowflakeType, tt.scale, got, tt.want)
			}
		})
	}
}

func TestConvertSQLValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  any
	}{
		{name: "nil value", value: nil, want: nil},
		{name: "string value", value: "hello", want: "hello"},
		{name: "int value", value: 42, want: 42},
		{name: "float value", value: 3.14, want: 3.14},
		{name: "bool value", value: true, want: true},

		// sql.Null types
		{
			name:  "NullString valid",
			value: sql.NullString{String: "test", Valid: true},
			want:  "test",
		},
		{
			name:  "NullString invalid",
			value: sql.NullString{String: "", Valid: false},
			want:  nil,
		},
		{
			name:  "NullInt64 valid",
			value: sql.NullInt64{Int64: 123, Valid: true},
			want:  int64(123),
		},
		{
			name:  "NullInt64 invalid",
			value: sql.NullInt64{Int64: 0, Valid: false},
			want:  nil,
		},
		{
			name:  "NullFloat64 valid",
			value: sql.NullFloat64{Float64: 1.5, Valid: true},
			want:  1.5,
		},
		{
			name:  "NullFloat64 invalid",
			value: sql.NullFloat64{Float64: 0, Valid: false},
			want:  nil,
		},
		{
			name:  "NullBool valid true",
			value: sql.NullBool{Bool: true, Valid: true},
			want:  true,
		},
		{
			name:  "NullBool valid false",
			value: sql.NullBool{Bool: false, Valid: true},
			want:  false,
		},
		{
			name:  "NullBool invalid",
			value: sql.NullBool{Bool: false, Valid: false},
			want:  nil,
		},
		{
			name:  "NullTime valid",
			value: sql.NullTime{Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Valid: true},
			want:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:  "NullTime invalid",
			value: sql.NullTime{Time: time.Time{}, Valid: false},
			want:  nil,
		},

		// Byte slices (JSON parsing)
		{
			name:  "bytes as JSON object",
			value: []byte(`{"key": "value"}`),
			want:  map[string]any{"key": "value"},
		},
		{
			name:  "bytes as JSON array",
			value: []byte(`[1, 2, 3]`),
			want:  []any{float64(1), float64(2), float64(3)},
		},
		{
			name:  "bytes as plain string (invalid JSON)",
			value: []byte("plain text"),
			want:  "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := convertSQLValue(tt.value)

			// Special handling for map comparison
			if wantMap, ok := tt.want.(map[string]any); ok {
				gotMap, ok := got.(map[string]any)
				if !ok {
					t.Errorf("convertSQLValue() = %T, want map[string]any", got)
					return
				}

				if len(gotMap) != len(wantMap) {
					t.Errorf("convertSQLValue() map length = %d, want %d", len(gotMap), len(wantMap))
					return
				}

				for k, v := range wantMap {
					if gotMap[k] != v {
						t.Errorf("convertSQLValue() map[%q] = %v, want %v", k, gotMap[k], v)
					}
				}

				return
			}

			// Special handling for slice comparison
			if wantSlice, ok := tt.want.([]any); ok {
				gotSlice, ok := got.([]any)
				if !ok {
					t.Errorf("convertSQLValue() = %T, want []any", got)
					return
				}

				if len(gotSlice) != len(wantSlice) {
					t.Errorf("convertSQLValue() slice length = %d, want %d", len(gotSlice), len(wantSlice))
					return
				}

				for i, v := range wantSlice {
					if gotSlice[i] != v {
						t.Errorf("convertSQLValue() slice[%d] = %v, want %v", i, gotSlice[i], v)
					}
				}

				return
			}

			// Special handling for time comparison
			if wantTime, ok := tt.want.(time.Time); ok {
				gotTime, ok := got.(time.Time)
				if !ok {
					t.Errorf("convertSQLValue() = %T, want time.Time", got)
					return
				}

				if !gotTime.Equal(wantTime) {
					t.Errorf("convertSQLValue() = %v, want %v", gotTime, wantTime)
				}

				return
			}

			if got != tt.want {
				t.Errorf("convertSQLValue() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	t.Parallel()

	truePtr := boolPtr(true)
	if truePtr == nil || *truePtr != true {
		t.Errorf("boolPtr(true) = %v, want pointer to true", truePtr)
	}

	falsePtr := boolPtr(false)
	if falsePtr == nil || *falsePtr != false {
		t.Errorf("boolPtr(false) = %v, want pointer to false", falsePtr)
	}

	// Ensure they're different pointers
	if truePtr == falsePtr {
		t.Errorf("boolPtr() returned same pointer for different values")
	}
}
