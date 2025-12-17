package snowflake

import (
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
)

func TestDetermineReadMode(t *testing.T) {
	t.Parallel()

	now := time.Now()
	hourAgo := now.Add(-time.Hour)

	tests := []struct {
		name   string
		params common.ReadParams
		want   readMode
	}{
		{
			name:   "no Since or Until - full backfill",
			params: common.ReadParams{ObjectName: "contacts"},
			want:   readModeFullBackfill,
		},
		{
			name: "only Since set - incremental from stream",
			params: common.ReadParams{
				ObjectName: "contacts",
				Since:      hourAgo,
			},
			want: readModeIncremental,
		},
		{
			name: "only Until set - time range from DT",
			params: common.ReadParams{
				ObjectName: "contacts",
				Until:      now,
			},
			want: readModeTimeRange,
		},
		{
			name: "both Since and Until set - time range from DT",
			params: common.ReadParams{
				ObjectName: "contacts",
				Since:      hourAgo,
				Until:      now,
			},
			want: readModeTimeRange,
		},
		{
			name: "zero time values treated as not set",
			params: common.ReadParams{
				ObjectName: "contacts",
				Since:      time.Time{},
				Until:      time.Time{},
			},
			want: readModeFullBackfill,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := determineReadMode(tt.params)
			if got != tt.want {
				t.Errorf("determineReadMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildStreamQuery(t *testing.T) {
	t.Parallel()

	c := &Connector{
		handle: &connectionInfo{
			database: "testdb",
			schema:   "testschema",
		},
	}

	tests := []struct {
		name       string
		streamName string
		objConfig  *objectConfig
		pageSize   int
		offset     int
		wantParts  []string // Parts that should be present in the query
		wantArgs   []any    // Expected args for prepared statement
	}{
		{
			name:       "basic stream query",
			streamName: `"TESTDB"."TESTSCHEMA"."CONTACTS_STREAM"`,
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey: "ID",
				},
			},
			pageSize: 100,
			offset:   0,
			wantParts: []string{
				"SELECT *",
				"METADATA$ACTION",
				"METADATA$ISUPDATE",
				"METADATA$ROW_ID",
				`FROM "TESTDB"."TESTSCHEMA"."CONTACTS_STREAM"`,
				`ORDER BY "ID" ASC`,
				"LIMIT ? OFFSET ?",
			},
			wantArgs: []any{100, 0},
		},
		{
			name:       "stream query with different primary key and offset",
			streamName: `"DB"."SCHEMA"."ORDERS_STREAM"`,
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey: "ORDER_ID",
				},
			},
			pageSize: 50,
			offset:   200,
			wantParts: []string{
				`ORDER BY "ORDER_ID" ASC`,
				"LIMIT ? OFFSET ?",
			},
			wantArgs: []any{50, 200},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, args := c.buildStreamQuery(tt.streamName, tt.objConfig, tt.pageSize, tt.offset)

			for _, part := range tt.wantParts {
				if !containsSubstring(got, part) {
					t.Errorf("buildStreamQuery() missing expected part %q in:\n%s", part, got)
				}
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("buildStreamQuery() args length = %d, want %d", len(args), len(tt.wantArgs))
			} else {
				for i, arg := range args {
					if arg != tt.wantArgs[i] {
						t.Errorf("buildStreamQuery() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestBuildDynamicTableQuery(t *testing.T) {
	t.Parallel()

	c := &Connector{
		handle: &connectionInfo{
			database: "testdb",
			schema:   "testschema",
		},
	}

	// Fixed time for consistent testing
	sinceTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	untilTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		tableName   string
		params      common.ReadParams
		objConfig   *objectConfig
		pageSize    int
		offset      int
		wantParts   []string // Parts that should be present
		unwantParts []string // Parts that should NOT be present
		wantArgs    []any    // Expected args for prepared statement
	}{
		{
			name:      "full backfill - no time filters",
			tableName: `"TESTDB"."TESTSCHEMA"."CONTACTS_DT"`,
			params:    common.ReadParams{ObjectName: "contacts"},
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey:      "ID",
					timestampColumn: "UPDATED_AT",
				},
			},
			pageSize: 100,
			offset:   0,
			wantParts: []string{
				"SELECT * FROM",
				`"TESTDB"."TESTSCHEMA"."CONTACTS_DT"`,
				`ORDER BY "ID" ASC`,
				"LIMIT ? OFFSET ?",
			},
			unwantParts: []string{
				"WHERE",
				"UPDATED_AT",
			},
			wantArgs: []any{100, 0},
		},
		{
			name:      "with Since filter",
			tableName: `"TESTDB"."TESTSCHEMA"."CONTACTS_DT"`,
			params: common.ReadParams{
				ObjectName: "contacts",
				Since:      sinceTime,
			},
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey:      "ID",
					timestampColumn: "UPDATED_AT",
				},
			},
			pageSize: 100,
			offset:   0,
			wantParts: []string{
				"WHERE",
				`"UPDATED_AT" >= ?`,
				`ORDER BY "ID" ASC`,
			},
			unwantParts: []string{
				"<=",
			},
			wantArgs: []any{"2024-01-15 10:00:00.000000", 100, 0},
		},
		{
			name:      "with Until filter",
			tableName: `"TESTDB"."TESTSCHEMA"."CONTACTS_DT"`,
			params: common.ReadParams{
				ObjectName: "contacts",
				Until:      untilTime,
			},
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey:      "ID",
					timestampColumn: "UPDATED_AT",
				},
			},
			pageSize: 100,
			offset:   0,
			wantParts: []string{
				"WHERE",
				`"UPDATED_AT" <= ?`,
			},
			unwantParts: []string{
				">=",
			},
			wantArgs: []any{"2024-01-15 12:00:00.000000", 100, 0},
		},
		{
			name:      "with both Since and Until",
			tableName: `"TESTDB"."TESTSCHEMA"."CONTACTS_DT"`,
			params: common.ReadParams{
				ObjectName: "contacts",
				Since:      sinceTime,
				Until:      untilTime,
			},
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey:      "ID",
					timestampColumn: "UPDATED_AT",
				},
			},
			pageSize: 100,
			offset:   0,
			wantParts: []string{
				"WHERE",
				`"UPDATED_AT" >= ?`,
				"AND",
				`"UPDATED_AT" <= ?`,
			},
			wantArgs: []any{"2024-01-15 10:00:00.000000", "2024-01-15 12:00:00.000000", 100, 0},
		},
		{
			name:      "with pagination offset",
			tableName: `"TESTDB"."TESTSCHEMA"."ORDERS_DT"`,
			params:    common.ReadParams{ObjectName: "orders"},
			objConfig: &objectConfig{
				dynamicTable: dynamicTableConfig{
					primaryKey: "ORDER_ID",
				},
			},
			pageSize: 50,
			offset:   150,
			wantParts: []string{
				`ORDER BY "ORDER_ID" ASC`,
				"LIMIT ? OFFSET ?",
			},
			wantArgs: []any{50, 150},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, args := c.buildDynamicTableQuery(tt.tableName, tt.params, tt.objConfig, tt.pageSize, tt.offset)

			for _, part := range tt.wantParts {
				if !containsSubstring(got, part) {
					t.Errorf("buildDynamicTableQuery() missing expected part %q in:\n%s", part, got)
				}
			}

			for _, part := range tt.unwantParts {
				if containsSubstring(got, part) {
					t.Errorf("buildDynamicTableQuery() should not contain %q in:\n%s", part, got)
				}
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("buildDynamicTableQuery() args length = %d, want %d", len(args), len(tt.wantArgs))
			} else {
				for i, arg := range args {
					if arg != tt.wantArgs[i] {
						t.Errorf("buildDynamicTableQuery() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestSnowflakeTimestampFormat(t *testing.T) {
	t.Parallel()

	// Verify the format string works correctly
	testTime := time.Date(2024, 6, 15, 14, 30, 45, 123456000, time.UTC)
	formatted := testTime.Format(SnowflakeTimestampFormat)

	expected := "2024-06-15 14:30:45.123456"
	if formatted != expected {
		t.Errorf("SnowflakeTimestampFormat produced %q, want %q", formatted, expected)
	}

	// Verify it handles midnight correctly
	midnight := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	formatted = midnight.Format(SnowflakeTimestampFormat)

	expected = "2024-01-01 00:00:00.000000"
	if formatted != expected {
		t.Errorf("SnowflakeTimestampFormat at midnight produced %q, want %q", formatted, expected)
	}
}

func TestDefaultPageSize(t *testing.T) {
	t.Parallel()

	if DefaultPageSize != 2000 {
		t.Errorf("DefaultPageSize = %d, want 2000", DefaultPageSize)
	}
}

func TestReadModeConstants(t *testing.T) {
	t.Parallel()

	// Ensure read modes are distinct
	modes := []readMode{readModeFullBackfill, readModeIncremental, readModeTimeRange}
	seen := make(map[readMode]bool)

	for _, mode := range modes {
		if seen[mode] {
			t.Errorf("duplicate readMode value: %d", mode)
		}

		seen[mode] = true
	}
}

// containsSubstring is a helper for checking if a string contains a substring.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
