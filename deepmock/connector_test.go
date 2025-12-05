package deepmock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test schemas used across multiple tests
var (
	testPersonSchema = []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"id": {"type": "string", "x-amp-id-field": true},
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0, "maximum": 150},
			"email": {"type": "string", "format": "email"},
			"updated": {"type": "integer", "x-amp-updated-field": true}
		},
		"required": ["name", "email"]
	}`)

	testProductSchema = []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"id": {"type": "integer", "x-amp-id-field": true},
			"name": {"type": "string"},
			"price": {"type": "number", "minimum": 0, "exclusiveMinimum": true},
			"category": {"type": "string", "enum": ["electronics", "clothing", "food"]},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"minItems": 1,
				"maxItems": 5,
				"uniqueItems": true
			},
			"lastModified": {"type": "string", "format": "date-time", "x-amp-updated-field": true}
		},
		"required": ["name", "category"]
	}`)

	testComplexSchema = []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"id": {"type": "string", "x-amp-id-field": true},
			"code": {"type": "string", "pattern": "^[A-Z]{3}-[0-9]{4}$"},
			"count": {"type": "integer", "multipleOf": 5},
			"nested": {
				"type": "object",
				"properties": {
					"field1": {"type": "string"},
					"field2": {"type": "number"}
				},
				"required": ["field1"]
			},
			"updated": {"type": "integer", "x-amp-updated-field": true}
		},
		"required": ["code"]
	}`)
)

// ============================================================================
// Constructor and Setup Tests
// ============================================================================

func TestNewConnector_Success(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons":  testPersonSchema,
		"products": testProductSchema,
	}

	conn, err := NewConnector(schemas)
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func TestNewConnector_EmptySchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schemas map[string][]byte
		wantErr bool
	}{
		{
			name:    "nil schemas",
			schemas: nil,
			wantErr: true,
		},
		{
			name:    "empty schemas",
			schemas: map[string][]byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn, err := NewConnector(tt.schemas)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, conn)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, conn)
			}
		})
	}
}

func TestNewConnector_InvalidSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schemas map[string][]byte
	}{
		{
			name: "malformed JSON",
			schemas: map[string][]byte{
				"invalid": []byte(`{"type": "object", "properties": {`),
			},
		},
		{
			name: "invalid schema structure",
			schemas: map[string][]byte{
				"invalid": []byte(`{"type": "invalid_type"}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn, err := NewConnector(tt.schemas)
			require.Error(t, err)
			assert.Nil(t, conn)
		})
	}
}

func TestNewConnector_WithOptions(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}

	// Test WithClient option
	conn, err := NewConnector(schemas, WithClient(http.DefaultClient))
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Test WithAuthenticatedClient option
	conn2, err := NewConnector(schemas, WithAuthenticatedClient(http.DefaultClient))
	require.NoError(t, err)
	require.NotNil(t, conn2)
}

// ============================================================================
// Schema Validation Tests
// ============================================================================

func TestWrite_ValidData(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	tests := []struct {
		name       string
		recordData map[string]any
	}{
		{
			name: "all fields valid",
			recordData: map[string]any{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
		},
		{
			name: "only required fields",
			recordData: map[string]any{
				"name":  "Jane Doe",
				"email": "jane@example.com",
			},
		},
		{
			name: "with explicit ID",
			recordData: map[string]any{
				"id":    "custom-id-123",
				"name":  "Bob Smith",
				"email": "bob@example.com",
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.Write(ctx, common.WriteParams{
				ObjectName: "persons",
				RecordId:   "",
				RecordData: tt.recordData,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, result.RecordId)
			assert.True(t, result.Success)
		})
	}
}

func TestWrite_InvalidData(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	tests := []struct {
		name       string
		recordData map[string]any
	}{
		{
			name: "missing required name",
			recordData: map[string]any{
				"email": "test@example.com",
			},
		},
		{
			name: "missing required email",
			recordData: map[string]any{
				"name": "Test User",
			},
		},
		{
			name: "wrong type for age",
			recordData: map[string]any{
				"name":  "Test User",
				"email": "test@example.com",
				"age":   "thirty",
			},
		},
		{
			name: "age out of range (too high)",
			recordData: map[string]any{
				"name":  "Test User",
				"email": "test@example.com",
				"age":   200,
			},
		},
		{
			name: "age out of range (negative)",
			recordData: map[string]any{
				"name":  "Test User",
				"email": "test@example.com",
				"age":   -5,
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.Write(ctx, common.WriteParams{
				ObjectName: "persons",
				RecordId:   "",
				RecordData: tt.recordData,
			})
			require.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestWrite_EnumValidation(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"products": testProductSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Valid enum value
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Laptop",
			"category": "electronics",
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Invalid enum value
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Book",
			"category": "books",
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestWrite_PatternValidation(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"complex": testComplexSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Valid pattern
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "complex",
		RecordData: map[string]any{
			"code": "ABC-1234",
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Invalid pattern
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "complex",
		RecordData: map[string]any{
			"code": "invalid-code",
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestWrite_NumericConstraints(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"products": testProductSchema,
		"complex":  testComplexSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Test exclusiveMinimum for price (must be > 0, not >= 0)
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Free Item",
			"category": "electronics",
			"price":    0,
		},
	})
	require.Error(t, err, "price should be exclusive minimum, 0 should fail")
	assert.Nil(t, result)

	// Test valid price > 0
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Paid Item",
			"category": "electronics",
			"price":    0.01,
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Test multipleOf constraint
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "complex",
		RecordData: map[string]any{
			"code":  "ABC-1234",
			"count": 10,
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Invalid multipleOf (not divisible by 5)
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "complex",
		RecordData: map[string]any{
			"code":  "ABC-1234",
			"count": 7,
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestWrite_ArrayConstraints(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"products": testProductSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Valid array (within bounds, unique items)
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Tagged Product",
			"category": "electronics",
			"tags":     []any{"tag1", "tag2", "tag3"},
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Too few items (minItems: 1)
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "No Tags",
			"category": "electronics",
			"tags":     []any{},
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)

	// Too many items (maxItems: 5)
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Too Many Tags",
			"category": "electronics",
			"tags":     []any{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"},
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)

	// Duplicate items (uniqueItems: true)
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":     "Duplicate Tags",
			"category": "electronics",
			"tags":     []any{"tag1", "tag1"},
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestWrite_NestedObjectValidation(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"complex": testComplexSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Valid nested object
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "complex",
		RecordData: map[string]any{
			"code": "ABC-1234",
			"nested": map[string]any{
				"field1": "value1",
				"field2": 42.5,
			},
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Missing required nested field
	result, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "complex",
		RecordData: map[string]any{
			"code": "ABC-1234",
			"nested": map[string]any{
				"field2": 42.5,
			},
		},
	})
	require.Error(t, err)
	assert.Nil(t, result)
}

// ============================================================================
// CRUD Operation Tests
// ============================================================================

func TestWrite_Create(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create new record
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
			"age":   25,
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.RecordId)
	assert.True(t, result.Success)

	// Verify record was created by reading it back
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("id", "name", "email", "age"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))
	assert.Equal(t, "Alice", readResult.Data[0].Fields["name"])
}

func TestWrite_Update(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create initial record
	createResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"name":  "Bob",
			"email": "bob@example.com",
			"age":   30,
		},
	})
	require.NoError(t, err)
	recordID := createResult.RecordId

	// Update the record
	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordId:   recordID,
		RecordData: map[string]any{
			"age": 31,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, recordID, updateResult.RecordId)
	assert.True(t, updateResult.Success)

	// Verify update (read back)
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("id", "name", "age"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))
	assert.Equal(t, "Bob", readResult.Data[0].Fields["name"])
	assert.Equal(t, float64(31), readResult.Data[0].Fields["age"])
}

func TestWrite_CreateWithExplicitID(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	customID := "my-custom-id"
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"id":    customID,
			"name":  "Charlie",
			"email": "charlie@example.com",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, customID, result.RecordId)

	// Verify ID was used
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("id", "name"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))
	assert.Equal(t, customID, readResult.Data[0].Fields["id"])
}

func TestRead_Basic(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create test records
	for i := 0; i < 3; i++ {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "persons",
			RecordData: map[string]any{
				"name":  fmt.Sprintf("Person %d", i),
				"email": fmt.Sprintf("person%d@example.com", i),
				"age":   20 + i,
			},
		})
		require.NoError(t, err)
	}

	// Read all records with field filtering
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name", "email"),
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Data))

	// Verify only requested fields are present
	for _, record := range result.Data {
		assert.Contains(t, record.Fields, "name")
		assert.Contains(t, record.Fields, "email")
		assert.NotContains(t, record.Fields, "age")
	}
}

func TestRead_Pagination(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create 25 records
	for i := 0; i < 25; i++ {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "persons",
			RecordData: map[string]any{
				"name":  fmt.Sprintf("Person %d", i),
				"email": fmt.Sprintf("person%d@example.com", i),
			},
		})
		require.NoError(t, err)
	}

	// First page
	page1, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
		NextPage:   "",
		PageSize:   10,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, len(page1.Data))
	assert.NotEmpty(t, page1.NextPage)
	assert.True(t, page1.Done == false)

	// Second page
	page2, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
		NextPage:   page1.NextPage,
		PageSize:   10,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, len(page2.Data))
	assert.NotEmpty(t, page2.NextPage)

	// Third page (final)
	page3, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
		NextPage:   page2.NextPage,
		PageSize:   10,
	})
	require.NoError(t, err)
	assert.Equal(t, 5, len(page3.Data))
	assert.Empty(t, page3.NextPage)
	assert.True(t, page3.Done)
}

func TestRead_TimeFiltering(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create records at different times
	baseTime := time.Now().Unix()

	// Record 1 (old)
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"name":    "Old Person",
			"email":   "old@example.com",
			"updated": baseTime - 100,
		},
	})
	require.NoError(t, err)

	// Record 2 (recent)
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"name":    "Recent Person",
			"email":   "recent@example.com",
			"updated": baseTime,
		},
	})
	require.NoError(t, err)

	// Read with Since filter (should get only recent)
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
		Since:      time.Unix(baseTime-50, 0),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Data))
	assert.Equal(t, "Recent Person", result.Data[0].Fields["name"])

	// Read with Until filter (should get only old)
	result, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
		Until:      time.Unix(baseTime-50, 0),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Data))
	assert.Equal(t, "Old Person", result.Data[0].Fields["name"])
}

func TestRead_EmptyResult(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Read from empty object
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
	})
	require.NoError(t, err)
	assert.Empty(t, result.Data)
	assert.True(t, result.Done)
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create record
	createResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"name":  "To Delete",
			"email": "delete@example.com",
		},
	})
	require.NoError(t, err)
	recordID := createResult.RecordId

	// Delete record
	deleteResult, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "persons",
		RecordId:   recordID,
	})
	require.NoError(t, err)
	assert.True(t, deleteResult.Success)

	// Verify deletion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
	})
	require.NoError(t, err)
	assert.Empty(t, readResult.Data)
}

func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Try to delete non-existent record
	result, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "persons",
		RecordId:   "non-existent-id",
	})
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons":  testPersonSchema,
		"products": testProductSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	result, err := conn.ListObjectMetadata(ctx, []string{"persons", "products"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Result))

	// Verify persons metadata
	personsMetadata := result.Result["persons"]
	assert.NotNil(t, personsMetadata)
	assert.Contains(t, personsMetadata.Fields, "name")
	assert.Contains(t, personsMetadata.Fields, "email")
	assert.Contains(t, personsMetadata.Fields, "age")

	// Verify required field
	nameField := personsMetadata.Fields["name"]
	assert.NotNil(t, nameField.IsRequired)
	assert.True(t, *nameField.IsRequired)

	// Verify products metadata
	productsMetadata := result.Result["products"]
	assert.NotNil(t, productsMetadata)
	categoryField := productsMetadata.Fields["category"]
	assert.NotNil(t, categoryField)
}

// ============================================================================
// Thread Safety Tests
// ============================================================================

func TestConcurrentWrites(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 100
	var wg sync.WaitGroup

	// Launch concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := conn.Write(ctx, common.WriteParams{
				ObjectName: "persons",
				RecordData: map[string]any{
					"name":  fmt.Sprintf("Person %d", index),
					"email": fmt.Sprintf("person%d@example.com", index),
				},
			})
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all records were created
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
	})
	require.NoError(t, err)
	assert.Equal(t, numGoroutines, len(result.Data))
}

func TestConcurrentReads(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create initial data
	for i := 0; i < 10; i++ {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "persons",
			RecordData: map[string]any{
				"name":  fmt.Sprintf("Person %d", i),
				"email": fmt.Sprintf("person%d@example.com", i),
			},
		})
		require.NoError(t, err)
	}

	// Launch concurrent reads
	numGoroutines := 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := conn.Read(ctx, common.ReadParams{
				ObjectName: "persons",
				Fields:     datautils.NewStringSet("name"),
			})
			assert.NoError(t, err)
			assert.Equal(t, 10, len(result.Data))
		}()
	}

	wg.Wait()
}

func TestConcurrentMixedOperations(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 50
	var wg sync.WaitGroup

	// Create initial records
	recordIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		result, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "persons",
			RecordData: map[string]any{
				"name":  fmt.Sprintf("Person %d", i),
				"email": fmt.Sprintf("person%d@example.com", i),
			},
		})
		require.NoError(t, err)
		recordIDs[i] = result.RecordId
	}

	// Mix of operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			switch index % 3 {
			case 0: // Write
				_, err := conn.Write(ctx, common.WriteParams{
					ObjectName: "persons",
					RecordData: map[string]any{
						"name":  fmt.Sprintf("New Person %d", index),
						"email": fmt.Sprintf("new%d@example.com", index),
					},
				})
				assert.NoError(t, err)

			case 1: // Read
				_, err := conn.Read(ctx, common.ReadParams{
					ObjectName: "persons",
					Fields:     datautils.NewStringSet("name"),
				})
				assert.NoError(t, err)

			case 2: // Update
				recordID := recordIDs[index%len(recordIDs)]
				_, err := conn.Write(ctx, common.WriteParams{
					ObjectName: "persons",
					RecordId:   recordID,
					RecordData: map[string]any{
						"age": index,
					},
				})
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()
}

// ============================================================================
// Data Isolation Tests
// ============================================================================

func TestMutationProtection_Read(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create record
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: map[string]any{
			"name":  "Original Name",
			"email": "original@example.com",
		},
	})
	require.NoError(t, err)

	// Read record
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name", "email"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))

	// Mutate returned data
	readResult.Data[0].Fields["name"] = "Mutated Name"
	readResult.Data[0].Fields["email"] = "mutated@example.com"

	// Read again to verify storage unchanged
	readResult2, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name", "email"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult2.Data))
	assert.Equal(t, "Original Name", readResult2.Data[0].Fields["name"])
	assert.Equal(t, "original@example.com", readResult2.Data[0].Fields["email"])
}

func TestMutationProtection_Write(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create mutable data
	data := map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
	}

	// Write record
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: data,
	})
	require.NoError(t, err)

	// Mutate original data after write
	data["name"] = "Mutated Alice"
	data["email"] = "mutated@example.com"

	// Read back to verify storage unchanged
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name", "email"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))
	assert.Equal(t, "Alice", readResult.Data[0].Fields["name"])
	assert.Equal(t, "alice@example.com", readResult.Data[0].Fields["email"])
	assert.Equal(t, result.RecordId, readResult.Data[0].Fields["id"])
}

func TestMutationProtection_GetAll(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create multiple records
	for i := 0; i < 3; i++ {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "persons",
			RecordData: map[string]any{
				"name":  fmt.Sprintf("Person %d", i),
				"email": fmt.Sprintf("person%d@example.com", i),
			},
		})
		require.NoError(t, err)
	}

	// Read all records
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name", "email"),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(readResult.Data))

	// Mutate the returned slice and maps
	readResult.Data[0].Fields["name"] = "Mutated"
	readResult.Data = append(readResult.Data, common.ReadResultRow{
		Fields: map[string]any{
			"name":  "Extra",
			"email": "extra@example.com",
		},
	})

	// Read again to verify storage unchanged
	readResult2, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(readResult2.Data))
	assert.Equal(t, "Person 0", readResult2.Data[0].Fields["name"])
}

func TestDeepCopyVerification(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Create record with nested data
	originalData := map[string]any{
		"name":  "Test",
		"email": "test@example.com",
	}

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "persons",
		RecordData: originalData,
	})
	require.NoError(t, err)

	// Read back
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name", "email"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))

	// Verify deep copy by checking pointer addresses differ
	// (If they were the same object, modifying one would affect the other)
	returnedData := readResult.Data[0]

	// Convert to JSON and back to ensure comparison works
	originalJSON, _ := json.Marshal(originalData)
	returnedJSON, _ := json.Marshal(returnedData)

	// Modify returned data
	returnedData.Fields["name"] = "Modified"

	// Re-read and verify original unchanged
	readResult2, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "persons",
		Fields:     datautils.NewStringSet("name"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Test", readResult2.Data[0].Fields["name"])

	_ = originalJSON
	_ = returnedJSON
	_ = result
}

// ============================================================================
// GenerateRandomRecord Tests
// ============================================================================

func TestGenerateRandomRecord_PrimitiveTypes(t *testing.T) {
	t.Parallel()

	primitiveSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"stringField": {"type": "string"},
			"intField": {"type": "integer"},
			"numberField": {"type": "number"},
			"boolField": {"type": "boolean"}
		}
	}`)

	schemas := map[string][]byte{
		"primitives": primitiveSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	record, err := conn.GenerateRandomRecord("primitives")
	require.NoError(t, err)
	require.NotNil(t, record)

	// Verify types
	_, hasString := record["stringField"].(string)
	assert.True(t, hasString)

	_, hasNumber := record["numberField"].(float64)
	assert.True(t, hasNumber)

	_, hasBool := record["boolField"].(bool)
	assert.True(t, hasBool)
}

func TestGenerateRandomRecord_Formats(t *testing.T) {
	t.Parallel()

	formatSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"email": {"type": "string", "format": "email"},
			"uuid": {"type": "string", "format": "uuid"},
			"date": {"type": "string", "format": "date"},
			"datetime": {"type": "string", "format": "date-time"},
			"phone": {"type": "string", "format": "phone"},
			"url": {"type": "string", "format": "uri"}
		}
	}`)

	schemas := map[string][]byte{
		"formats": formatSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	record, err := conn.GenerateRandomRecord("formats")
	require.NoError(t, err)
	require.NotNil(t, record)

	// Verify formats (basic checks)
	email, ok := record["email"].(string)
	assert.True(t, ok)
	assert.Contains(t, email, "@")

	uuid, ok := record["uuid"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, uuid)

	url, ok := record["url"].(string)
	assert.True(t, ok)
	assert.Contains(t, url, "http")
}

func TestGenerateRandomRecord_Enums(t *testing.T) {
	t.Parallel()

	enumSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"status": {"type": "string", "enum": ["active", "inactive", "pending"]}
		}
	}`)

	schemas := map[string][]byte{
		"enums": enumSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	// Generate multiple records to test randomness
	validValues := map[string]bool{"active": true, "inactive": true, "pending": true}

	for i := 0; i < 10; i++ {
		record, err := conn.GenerateRandomRecord("enums")
	require.NoError(t, err)
		require.NotNil(t, record)

		status, ok := record["status"].(string)
		assert.True(t, ok)
		assert.True(t, validValues[status], "status should be one of the enum values")
	}
}

func TestGenerateRandomRecord_NumericConstraints(t *testing.T) {
	t.Parallel()

	constraintSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"age": {"type": "integer", "minimum": 18, "maximum": 65},
			"price": {"type": "number", "minimum": 0, "exclusiveMinimum": true, "maximum": 1000},
			"count": {"type": "integer", "multipleOf": 5}
		}
	}`)

	schemas := map[string][]byte{
		"constraints": constraintSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	// Generate multiple records to test constraints
	for i := 0; i < 20; i++ {
		record, err := conn.GenerateRandomRecord("constraints")
	require.NoError(t, err)
		require.NotNil(t, record)

		if age, ok := record["age"].(float64); ok {
			assert.GreaterOrEqual(t, age, float64(18))
			assert.LessOrEqual(t, age, float64(65))
		}

		if price, ok := record["price"].(float64); ok {
			assert.Greater(t, price, float64(0))
			assert.LessOrEqual(t, price, float64(1000))
		}

		if count, ok := record["count"].(float64); ok {
			assert.Equal(t, float64(0), float64(int(count)%5))
		}
	}
}

func TestGenerateRandomRecord_StringConstraints(t *testing.T) {
	t.Parallel()

	stringSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"shortCode": {"type": "string", "minLength": 3, "maxLength": 5},
			"pattern": {"type": "string", "pattern": "^[A-Z]{2}-[0-9]{3}$"}
		}
	}`)

	schemas := map[string][]byte{
		"strings": stringSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		record, err := conn.GenerateRandomRecord("strings")
	require.NoError(t, err)
		require.NotNil(t, record)

		if shortCode, ok := record["shortCode"].(string); ok {
			assert.GreaterOrEqual(t, len(shortCode), 3)
			assert.LessOrEqual(t, len(shortCode), 5)
		}

		if pattern, ok := record["pattern"].(string); ok {
			assert.Regexp(t, `^[A-Z]{2}-[0-9]{3}$`, pattern)
		}
	}
}

func TestGenerateRandomRecord_Arrays(t *testing.T) {
	t.Parallel()

	arraySchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"minItems": 2,
				"maxItems": 5,
				"uniqueItems": true
			}
		}
	}`)

	schemas := map[string][]byte{
		"arrays": arraySchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		record, err := conn.GenerateRandomRecord("arrays")
	require.NoError(t, err)
		require.NotNil(t, record)

		if tags, ok := record["tags"].([]any); ok {
			assert.GreaterOrEqual(t, len(tags), 2)
			assert.LessOrEqual(t, len(tags), 5)

			// Verify uniqueness
			seen := make(map[string]bool)
			for _, tag := range tags {
				tagStr := tag.(string)
				assert.False(t, seen[tagStr], "tags should be unique")
				seen[tagStr] = true
			}
		}
	}
}

func TestGenerateRandomRecord_NestedObjects(t *testing.T) {
	t.Parallel()

	nestedSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer"}
				},
				"required": ["name"]
			}
		}
	}`)

	schemas := map[string][]byte{
		"nested": nestedSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	record, err := conn.GenerateRandomRecord("nested")
	require.NoError(t, err)
	require.NotNil(t, record)

	if user, ok := record["user"].(map[string]any); ok {
		assert.Contains(t, user, "name")
		_, hasName := user["name"].(string)
		assert.True(t, hasName)
	}
}

func TestGenerateRandomRecord_DepthLimit(t *testing.T) {
	t.Parallel()

	// Deeply nested schema that would recurse infinitely without depth limit
	deepSchema := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"level1": {
				"type": "object",
				"properties": {
					"level2": {
						"type": "object",
						"properties": {
							"level3": {
								"type": "object",
								"properties": {
									"level4": {
										"type": "object",
										"properties": {
											"level5": {
												"type": "object",
												"properties": {
													"level6": {"type": "string"}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`)

	schemas := map[string][]byte{
		"deep": deepSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	// Should not panic or hang
	record, err := conn.GenerateRandomRecord("deep")
	require.NoError(t, err)
	require.NotNil(t, record)
}

func TestGenerateRandomRecord_SpecialFields(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	record, err := conn.GenerateRandomRecord("persons")
	require.NoError(t, err)
	require.NotNil(t, record)

	// Verify ID field is generated
	assert.Contains(t, record, "id")
	id, ok := record["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)

	// Verify updated field is generated (timestamp)
	assert.Contains(t, record, "updated")
	updated, ok := record["updated"].(float64)
	assert.True(t, ok)
	assert.Greater(t, updated, float64(0))
}

func TestGenerateRandomRecord_ValidationSuccess(t *testing.T) {
	t.Parallel()

	schemas := map[string][]byte{
		"persons":  testPersonSchema,
		"products": testProductSchema,
		"complex":  testComplexSchema,
	}
	conn, err := NewConnector(schemas)
	require.NoError(t, err)

	ctx := context.Background()

	// Test that generated records pass validation when written
	testCases := []string{"persons", "products", "complex"}

	for _, objectName := range testCases {
		t.Run(objectName, func(t *testing.T) {
			// Generate multiple random records
			for i := 0; i < 5; i++ {
				record, err := conn.GenerateRandomRecord(objectName)
	require.NoError(t, err)
				require.NotNil(t, record)

				// Try to write the generated record (validates schema)
				result, err := conn.Write(ctx, common.WriteParams{
					ObjectName: objectName,
					RecordData: record,
				})
				require.NoError(t, err, "generated record should pass validation")
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.RecordId)
			}
		})
	}
}
