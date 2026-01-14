package snowflake

import (
	"testing"
)

func TestNewSnowflakeObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		paramsMap map[string]string
		want      Objects
		wantErr   bool
	}{
		{
			name: "parses query in dynamicTable",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable']['query']": "SELECT * FROM customers",
			},
			want: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query: "SELECT * FROM customers",
					},
				},
			},
		},
		{
			name: "parses dynamicTable nested properties",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable']['primaryKey']":      "id",
				"$['objects']['contacts']['dynamicTable']['timestampColumn']": "updated_at",
				"$['objects']['contacts']['dynamicTable']['targetLag']":       "1 hour",
				"$['objects']['contacts']['dynamicTable']['name']":            "contacts_dt",
			},
			want: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						primaryKey:      "id",
						timestampColumn: "updated_at",
						targetLag:       "1 hour",
						name:            "contacts_dt",
					},
				},
			},
		},
		{
			name: "parses stream nested properties",
			paramsMap: map[string]string{
				"$['objects']['contacts']['stream']['name']": "contacts_stream",
			},
			want: Objects{
				"contacts": {
					stream: streamConfig{
						name: "contacts_stream",
					},
				},
			},
		},
		{
			name: "parses full object config",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable']['query']":           "SELECT * FROM customers",
				"$['objects']['contacts']['dynamicTable']['primaryKey']":      "id",
				"$['objects']['contacts']['dynamicTable']['timestampColumn']": "updated_at",
				"$['objects']['contacts']['dynamicTable']['targetLag']":       "1 hour",
				"$['objects']['contacts']['dynamicTable']['name']":            "contacts_dt",
				"$['objects']['contacts']['stream']['name']":                  "contacts_stream",
			},
			want: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:           "SELECT * FROM customers",
						primaryKey:      "id",
						timestampColumn: "updated_at",
						targetLag:       "1 hour",
						name:            "contacts_dt",
					},
					stream: streamConfig{
						name: "contacts_stream",
					},
				},
			},
		},
		{
			name: "parses multiple objects",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable']['query']":      "SELECT * FROM customers",
				"$['objects']['contacts']['dynamicTable']['primaryKey']": "id",
				"$['objects']['orders']['dynamicTable']['query']":        "SELECT * FROM orders",
				"$['objects']['orders']['dynamicTable']['primaryKey']":   "order_id",
			},
			want: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:      "SELECT * FROM customers",
						primaryKey: "id",
					},
				},
				"orders": {
					dynamicTable: dynamicTableConfig{
						query:      "SELECT * FROM orders",
						primaryKey: "order_id",
					},
				},
			},
		},
		{
			name: "ignores non-object keys",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable']['query']":      "SELECT * FROM customers",
				"$['objects']['contacts']['dynamicTable']['primaryKey']": "id",
				"$['other']['key']['value']['test']":                     "value",
				"someOtherKey":                                           "value",
			},
			want: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:      "SELECT * FROM customers",
						primaryKey: "id",
					},
				},
			},
		},
		{
			name:      "handles empty paramsMap",
			paramsMap: map[string]string{},
			want:      Objects{},
		},
		{
			name: "returns error for invalid path syntax",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable'][\"query\"]": "SELECT *",
			},
			wantErr: true,
		},
		{
			name: "returns error for wrong path depth (3 levels)",
			paramsMap: map[string]string{
				"$['objects']['contacts']['query']": "SELECT *",
			},
			wantErr: true,
		},
		{
			name: "returns error for invalid parent key",
			paramsMap: map[string]string{
				"$['objects']['contacts']['invalid']['query']": "SELECT *",
			},
			wantErr: true,
		},
		{
			name: "returns error for unknown dynamicTable property",
			paramsMap: map[string]string{
				"$['objects']['contacts']['dynamicTable']['unknownProperty']": "value",
			},
			wantErr: true,
		},
		{
			name: "returns error for unknown stream property",
			paramsMap: map[string]string{
				"$['objects']['contacts']['stream']['unknownProperty']": "value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := newSnowflakeObjects(tt.paramsMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("newSnowflakeObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(*got) != len(tt.want) {
				t.Errorf("newSnowflakeObjects() got %d objects, want %d", len(*got), len(tt.want))
				return
			}

			for objName, wantCfg := range tt.want {
				gotCfg, ok := (*got)[objName]
				if !ok {
					t.Errorf("newSnowflakeObjects() missing object %q", objName)
					continue
				}

				if gotCfg.dynamicTable.query != wantCfg.dynamicTable.query {
					t.Errorf("object %q: dynamicTable.query = %q, want %q",
						objName, gotCfg.dynamicTable.query, wantCfg.dynamicTable.query)
				}

				if gotCfg.dynamicTable.primaryKey != wantCfg.dynamicTable.primaryKey {
					t.Errorf("object %q: dynamicTable.primaryKey = %q, want %q",
						objName, gotCfg.dynamicTable.primaryKey, wantCfg.dynamicTable.primaryKey)
				}

				if gotCfg.dynamicTable.timestampColumn != wantCfg.dynamicTable.timestampColumn {
					t.Errorf("object %q: dynamicTable.timestampColumn = %q, want %q",
						objName, gotCfg.dynamicTable.timestampColumn, wantCfg.dynamicTable.timestampColumn)
				}

				if gotCfg.dynamicTable.targetLag != wantCfg.dynamicTable.targetLag {
					t.Errorf("object %q: dynamicTable.targetLag = %q, want %q",
						objName, gotCfg.dynamicTable.targetLag, wantCfg.dynamicTable.targetLag)
				}

				if gotCfg.dynamicTable.name != wantCfg.dynamicTable.name {
					t.Errorf("object %q: dynamicTable.name = %q, want %q",
						objName, gotCfg.dynamicTable.name, wantCfg.dynamicTable.name)
				}

				if gotCfg.stream.name != wantCfg.stream.name {
					t.Errorf("object %q: stream.name = %q, want %q",
						objName, gotCfg.stream.name, wantCfg.stream.name)
				}
			}
		})
	}
}

func TestObjects_ToMetadataMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		objects Objects
		want    map[string]string
	}{
		{
			name: "converts full object config to metadata map",
			objects: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:           "SELECT * FROM customers",
						primaryKey:      "id",
						timestampColumn: "updated_at",
						targetLag:       "1 hour",
						name:            "contacts_dt",
					},
					stream: streamConfig{
						name: "contacts_stream",
					},
				},
			},
			want: map[string]string{
				"$['objects']['contacts']['dynamicTable']['query']":           "SELECT * FROM customers",
				"$['objects']['contacts']['dynamicTable']['primaryKey']":      "id",
				"$['objects']['contacts']['dynamicTable']['timestampColumn']": "updated_at",
				"$['objects']['contacts']['dynamicTable']['targetLag']":       "1 hour",
				"$['objects']['contacts']['dynamicTable']['name']":            "contacts_dt",
				"$['objects']['contacts']['stream']['name']":                  "contacts_stream",
			},
		},
		{
			name: "omits empty values",
			objects: Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:      "SELECT * FROM customers",
						primaryKey: "id",
					},
				},
			},
			want: map[string]string{
				"$['objects']['contacts']['dynamicTable']['query']":      "SELECT * FROM customers",
				"$['objects']['contacts']['dynamicTable']['primaryKey']": "id",
			},
		},
		{
			name:    "handles empty objects",
			objects: Objects{},
			want:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.objects.ToMetadataMap()

			if len(got) != len(tt.want) {
				t.Errorf("ToMetadataMap() got %d entries, want %d", len(got), len(tt.want))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.want)

				return
			}

			for key, wantVal := range tt.want {
				gotVal, ok := got[key]
				if !ok {
					t.Errorf("ToMetadataMap() missing key %q", key)
					continue
				}

				if gotVal != wantVal {
					t.Errorf("ToMetadataMap() key %q = %q, want %q", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestObjects_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		objects *Objects
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil objects is valid",
			objects: nil,
			wantErr: false,
		},
		{
			name:    "empty objects is valid",
			objects: &Objects{},
			wantErr: false,
		},
		{
			name: "valid object with query and primaryKey",
			objects: &Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:      "SELECT * FROM customers",
						primaryKey: "id",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing query",
			objects: &Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						primaryKey: "id",
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'dynamicTable.query'",
		},
		{
			name: "missing primaryKey",
			objects: &Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query: "SELECT * FROM customers",
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'dynamicTable.primaryKey'",
		},
		{
			name: "missing both query and primaryKey",
			objects: &Objects{
				"contacts": {},
			},
			wantErr: true,
			errMsg:  "query",
		},
		{
			name: "multiple objects with one invalid",
			objects: &Objects{
				"contacts": {
					dynamicTable: dynamicTableConfig{
						query:      "SELECT * FROM customers",
						primaryKey: "id",
					},
				},
				"orders": {
					dynamicTable: dynamicTableConfig{
						query: "SELECT * FROM orders",
						// missing primaryKey
					},
				},
			},
			wantErr: true,
			errMsg:  "orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.objects.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestObjects_Get(t *testing.T) {
	t.Parallel()

	objects := Objects{
		"contacts": {
			dynamicTable: dynamicTableConfig{
				query:      "SELECT * FROM customers",
				primaryKey: "id",
			},
		},
	}

	t.Run("returns existing object", func(t *testing.T) {
		t.Parallel()

		cfg, ok := objects.Get("contacts")
		if !ok {
			t.Error("Get() returned ok = false for existing object")
			return
		}

		if cfg.dynamicTable.query != "SELECT * FROM customers" {
			t.Errorf("Get() query = %q, want %q", cfg.dynamicTable.query, "SELECT * FROM customers")
		}
	})

	t.Run("returns false for non-existing object", func(t *testing.T) {
		t.Parallel()

		_, ok := objects.Get("nonexistent")
		if ok {
			t.Error("Get() returned ok = true for non-existing object")
		}
	})
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	// Test that parsing and ToMetadataMap are inverses of each other
	original := map[string]string{
		"$['objects']['contacts']['dynamicTable']['query']":           "SELECT * FROM customers",
		"$['objects']['contacts']['dynamicTable']['primaryKey']":      "id",
		"$['objects']['contacts']['dynamicTable']['timestampColumn']": "updated_at",
		"$['objects']['contacts']['dynamicTable']['targetLag']":       "1 hour",
		"$['objects']['contacts']['dynamicTable']['name']":            "contacts_dt",
		"$['objects']['contacts']['stream']['name']":                  "contacts_stream",
	}

	// Parse the metadata map
	objects, err := newSnowflakeObjects(original)
	if err != nil {
		t.Fatalf("newSnowflakeObjects() error = %v", err)
	}

	// Convert back to metadata map
	roundTripped := objects.ToMetadataMap()

	// Verify all original keys are present
	for key, wantVal := range original {
		gotVal, ok := roundTripped[key]
		if !ok {
			t.Errorf("round trip missing key %q", key)
			continue
		}

		if gotVal != wantVal {
			t.Errorf("round trip key %q = %q, want %q", key, gotVal, wantVal)
		}
	}

	// Verify no extra keys
	if len(roundTripped) != len(original) {
		t.Errorf("round trip has %d keys, want %d", len(roundTripped), len(original))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
