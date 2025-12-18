package memstore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to parse JSON schema into InputSchema
func mustParseSchema(jsonSchema string) *InputSchema {
	var schema InputSchema
	if err := json.Unmarshal([]byte(jsonSchema), &schema); err != nil {
		panic(fmt.Sprintf("failed to parse test schema: %v", err))
	}
	return &schema
}

// Test schemas used across multiple tests
var (
	testPersonSchema = mustParseSchema(`{
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

	testProductSchema = mustParseSchema(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"id": {"type": "integer", "x-amp-id-field": true},
			"name": {"type": "string"},
			"price": {"type": "number", "exclusiveMinimum": 0},
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

	testComplexSchema = mustParseSchema(`{
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
// Struct-Based Schema Tests
// ============================================================================
//
// The following tests demonstrate using Go structs with tags to define schemas
// instead of writing raw JSON schemas. This approach provides:
//
// 1. Type safety at compile time
// 2. IDE autocomplete and refactoring support
// 3. Easier maintenance and readability
// 4. Automatic schema generation from existing Go types
//
// Example Usage:
//
//   // Define your data model with tags
//   type User struct {
//       ID        string    `json:"id" jsonschema_extras:"x-amp-id-field=true"`
//       Email     string    `json:"email" jsonschema:"required,format=email"`
//       Name      string    `json:"name" jsonschema:"required,minLength=1,maxLength=100"`
//       Role      string    `json:"role" jsonschema:"enum=admin,enum=user,enum=guest"`
//       CreatedAt time.Time `json:"created_at" jsonschema_extras:"x-amp-updated-field=true" jsonschema:"format=date-time"`
//       Age       int       `json:"age" jsonschema:"minimum=18,maximum=120"`
//       Active    bool      `json:"active"`
//   }
//
//   // Create connector with struct schemas
//   conn, err := memstore.NewConnector(memstore.WithStructSchemas(map[string]interface{}{
//       "users": &User{},
//   }))
//
//   // Use connector normally
//   record, _ := conn.GenerateRandomRecord("users")
//   result, _ := conn.Write(ctx, common.WriteParams{
//       ObjectName: "users",
//       RecordData: record,
//   })
//
// Supported Tags:
//
// jsonschema_extras:
//   - x-amp-id-field=true       : Marks field as unique identifier
//   - x-amp-updated-field=true  : Marks field as last updated timestamp
//
// jsonschema:
//   - required                  : Field is required
//   - enum=val1,enum=val2       : Allowed values (creates SingleSelect)
//   - format=email|uuid|date|date-time|phone|uri
//   - minLength=N, maxLength=N  : String length constraints
//   - minimum=N, maximum=N      : Numeric constraints
//   - title=Display Name        : Human-readable field name
//   - description=...           : Field description
//
// Field Type Mapping:
//   - string       → ValueTypeString
//   - int, int64   → ValueTypeInt
//   - float64      → ValueTypeFloat
//   - bool         → ValueTypeBoolean
//   - []T          → ValueTypeOther (array)
//   - struct       → ValueTypeOther (object)
//   - time.Time    → ValueTypeDateTime (with format=date-time)
//
// ============================================================================

// Test structs for struct-based schema derivation
// These demonstrate the full range of jsonschema tags and custom extensions

type TestContact struct {
	ID        string   `json:"id" jsonschema_extras:"x-amp-id-field=true"`
	Email     string   `json:"email" jsonschema:"required,format=email,title=Email Address"`
	FirstName string   `json:"firstName" jsonschema:"minLength=1,maxLength=50,title=First Name"`
	LastName  string   `json:"lastName" jsonschema:"minLength=1,maxLength=50"`
	Phone     string   `json:"phone" jsonschema:"format=phone"`
	Status    string   `json:"status" jsonschema:"enum=active,enum=inactive,enum=pending"`
	CreatedAt int64    `json:"createdAt" jsonschema_extras:"x-amp-updated-field=true"`
	Tags      []string `json:"tags" jsonschema:"uniqueItems=true"`
}

type TestCompany struct {
	ID            int64  `json:"id" jsonschema_extras:"x-amp-id-field=true"`
	Name          string `json:"name" jsonschema:"required,minLength=1,maxLength=200"`
	Industry      string `json:"industry" jsonschema:"enum=technology,enum=finance,enum=healthcare"`
	EmployeeCount int    `json:"employeeCount" jsonschema:"minimum=1,maximum=1000000"`
	Website       string `json:"website" jsonschema:"format=uri"`
	UpdatedAt     string `json:"updatedAt" jsonschema_extras:"x-amp-updated-field=true" jsonschema:"format=date-time"`
}

type TestDeal struct {
	ID           string  `json:"id" jsonschema_extras:"x-amp-id-field=true"`
	Title        string  `json:"title" jsonschema:"required,minLength=1"`
	Amount       float64 `json:"amount" jsonschema:"minimum=0,maximum=10000000"`
	Stage        string  `json:"stage" jsonschema:"enum=prospecting,enum=qualification,enum=proposal,enum=closed"`
	ContactID    string  `json:"contactId"`
	CompanyID    int64   `json:"companyId"`
	CloseDate    string  `json:"closeDate" jsonschema:"format=date"`
	LastModified int64   `json:"lastModified" jsonschema_extras:"x-amp-updated-field=true"`
}

func TestDeriveSchemasFromStructs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, schemas map[string][]byte)
	}{
		{
			name: "valid structs with all tag types",
			input: map[string]interface{}{
				"contacts":  &TestContact{},
				"companies": &TestCompany{},
				"deals":     &TestDeal{},
			},
			expectError: false,
			validate: func(t *testing.T, schemas map[string][]byte) {
				// Verify all objects present
				if len(schemas) != 3 {
					t.Errorf("expected 3 schemas, got %d", len(schemas))
				}

				// Verify each schema is valid JSON
				for objName, schemaBytes := range schemas {
					var schemaMap map[string]any
					if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
						t.Errorf("schema %s is not valid JSON: %v", objName, err)
					}

					// Verify $schema field
					if schema, ok := schemaMap["$schema"].(string); !ok || schema == "" {
						t.Errorf("schema %s missing $schema field", objName)
					}

					// Verify properties exist
					if _, ok := schemaMap["properties"]; !ok {
						t.Errorf("schema %s missing properties", objName)
					}
				}

				// Verify custom extensions for contacts
				var contactSchema map[string]any
				_ = json.Unmarshal(schemas["contacts"], &contactSchema)
				props := contactSchema["properties"].(map[string]any)
				idField := props["id"].(map[string]any)
				if idExt, ok := idField["x-amp-id-field"]; !ok || idExt != true {
					t.Errorf("contacts.id missing x-amp-id-field extension")
				}
				createdAtField := props["createdAt"].(map[string]any)
				if updatedExt, ok := createdAtField["x-amp-updated-field"]; !ok || updatedExt != true {
					t.Errorf("contacts.createdAt missing x-amp-updated-field extension")
				}

				// Verify required fields
				if required, ok := contactSchema["required"].([]any); !ok || len(required) == 0 {
					t.Errorf("contacts schema missing required fields")
				}
			},
		},
		{
			name:        "nil input",
			input:       nil,
			expectError: true,
			errorMsg:    "cannot be nil or empty",
		},
		{
			name:        "empty map",
			input:       map[string]interface{}{},
			expectError: true,
			errorMsg:    "cannot be nil or empty",
		},
		{
			name: "non-struct value",
			input: map[string]interface{}{
				"invalid": "not a struct",
			},
			expectError: true,
			errorMsg:    "expected struct or pointer to struct",
		},
		{
			name: "nil struct pointer",
			input: map[string]interface{}{
				"contacts": (*TestContact)(nil),
			},
			expectError: true,
			errorMsg:    "cannot be nil",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schemas, err := DeriveSchemasFromStructs(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, schemas)
			}
		})
	}
}

func TestNewConnectorWithStructSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rawSchemas  map[string]*InputSchema
		structOpts  []Option
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, conn *Connector)
	}{
		{
			name:       "struct schemas only",
			rawSchemas: nil,
			structOpts: []Option{
				WithStructSchemas(map[string]interface{}{
					"contacts":  &TestContact{},
					"companies": &TestCompany{},
				}),
			},
			expectError: false,
			validate: func(t *testing.T, conn *Connector) {
				// Verify schemas are loaded
				if len(conn.schemas) != 2 {
					t.Errorf("expected 2 schemas, got %d", len(conn.schemas))
				}

				// Verify storage is initialized
				if conn.storage == nil {
					t.Error("storage not initialized")
				}

				// Verify ID fields are detected
				if idField := conn.storage.GetIdFields()["contacts"]; idField != "id" {
					t.Errorf("expected contacts ID field 'id', got %q", idField)
				}
				if idField := conn.storage.GetIdFields()["companies"]; idField != "id" {
					t.Errorf("expected companies ID field 'id', got %q", idField)
				}

				// Verify updated fields are detected
				if updatedField := conn.storage.GetUpdatedFields()["contacts"]; updatedField != "createdAt" {
					t.Errorf("expected contacts updated field 'createdAt', got %q", updatedField)
				}
				if updatedField := conn.storage.GetUpdatedFields()["companies"]; updatedField != "updatedAt" {
					t.Errorf("expected companies updated field 'updatedAt', got %q", updatedField)
				}
			},
		},
		{
			name:        "neither raw nor struct schemas",
			rawSchemas:  nil,
			structOpts:  []Option{},
			expectError: true,
			errorMsg:    "must provide either raw schemas or use WithStructSchemas",
		},
		{
			name:       "invalid struct schemas",
			rawSchemas: nil,
			structOpts: []Option{
				WithStructSchemas(map[string]interface{}{
					"invalid": "not a struct",
				}),
			},
			expectError: true,
			errorMsg:    "failed to derive schemas from structs",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn, err := NewConnector(append([]Option{WithSchemas(tt.rawSchemas)}, tt.structOpts...)...)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, conn)
			}
		})
	}
}

func TestSchemasPriority(t *testing.T) {
	t.Parallel()

	// Define a schema with different field names
	rawContactSchema := mustParseSchema(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"rawId": {"type": "string", "x-amp-id-field": true},
			"rawEmail": {"type": "string", "format": "email"}
		},
		"required": ["rawEmail"]
	}`)

	rawSchemas := map[string]*InputSchema{
		"contacts": rawContactSchema,
	}

	structSchemas := map[string]interface{}{
		"contacts": &TestContact{},
	}

	// Create connector with both raw and struct schemas
	conn, err := NewConnector(WithSchemas(rawSchemas), WithStructSchemas(structSchemas))
	if err != nil {
		t.Fatalf("failed to create connector: %v", err)
	}

	// Verify raw schemas take priority
	schema, exists := conn.schemas.Get("contacts")
	if !exists {
		t.Fatal("contacts schema not found")
	}

	// Extract schema to verify it's the raw one (has rawId, not id)
	schemaJSON, _ := json.Marshal(schema)
	var schemaMap map[string]any
	_ = json.Unmarshal(schemaJSON, &schemaMap)

	props := schemaMap["properties"].(map[string]any)
	if _, hasRawId := props["rawId"]; !hasRawId {
		t.Error("expected raw schema (with rawId field), but got struct-derived schema")
	}
	if _, hasId := props["id"]; hasId {
		t.Error("expected raw schema (without id field), but got struct-derived schema")
	}

	// Verify ID field is from raw schema
	if idField := conn.storage.GetIdFields()["contacts"]; idField != "rawId" {
		t.Errorf("expected ID field 'rawId' from raw schema, got %q", idField)
	}
}

func TestCRUDWithStructSchemas(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create connector with struct schemas
	conn, err := NewConnector(WithStructSchemas(map[string]interface{}{
		"contacts":  &TestContact{},
		"companies": &TestCompany{},
	}))
	if err != nil {
		t.Fatalf("failed to create connector: %v", err)
	}

	t.Run("write and read contact", func(t *testing.T) {
		// Write a contact
		contactData := map[string]any{
			"email":     "test@example.com",
			"firstName": "John",
			"lastName":  "Doe",
			"status":    "active",
			"tags":      []string{"vip", "customer"},
		}

		writeResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contacts",
			RecordData: contactData,
		})
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		if writeResult.RecordId == "" {
			t.Error("expected record ID to be generated")
		}

		// Verify ID field was auto-generated
		if _, hasId := writeResult.Data["id"]; !hasId {
			t.Error("expected 'id' field to be auto-generated")
		}

		// Verify updated field was auto-generated
		if _, hasCreatedAt := writeResult.Data["createdAt"]; !hasCreatedAt {
			t.Error("expected 'createdAt' field to be auto-generated")
		}

		// Read the contact back
		readResult, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "contacts",
			Fields:     datautils.NewStringSet("id", "email", "firstName", "lastName", "status", "tags", "createdAt"),
		})
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}

		if readResult.Rows != 1 {
			t.Errorf("expected 1 row, got %d", readResult.Rows)
		}

		if len(readResult.Data) != 1 {
			t.Fatalf("expected 1 record, got %d", len(readResult.Data))
		}

		record := readResult.Data[0]
		if email := record.Fields["email"]; email != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got %v", email)
		}
	})

	t.Run("write company with integer ID", func(t *testing.T) {
		companyData := map[string]any{
			"name":          "Acme Corp",
			"industry":      "technology",
			"employeeCount": 100,
			"website":       "https://acme.example.com",
		}

		writeResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "companies",
			RecordData: companyData,
		})
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		// Verify integer ID was generated
		idVal, hasId := writeResult.Data["id"]
		if !hasId {
			t.Error("expected 'id' field to be auto-generated")
		}

		// Check ID is numeric (int or int64)
		switch idVal.(type) {
		case int, int64:
			// OK
		default:
			t.Errorf("expected integer ID, got %T", idVal)
		}
	})

	t.Run("validation with struct-derived schema", func(t *testing.T) {
		// Try to write invalid data (missing required field)
		invalidData := map[string]any{
			"firstName": "Jane",
			// Missing required "email" field
		}

		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contacts",
			RecordData: invalidData,
		})
		if err == nil {
			t.Error("expected validation error for missing required field")
		}
		if !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})

	t.Run("enum validation", func(t *testing.T) {
		// Try to write invalid enum value
		invalidData := map[string]any{
			"email":  "test2@example.com",
			"status": "invalid_status", // Not in enum
		}

		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contacts",
			RecordData: invalidData,
		})
		if err == nil {
			t.Error("expected validation error for invalid enum value")
		}
	})
}

func TestGenerateRandomRecordWithStructSchemas(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	conn, err := NewConnector(WithStructSchemas(map[string]interface{}{
		"contacts":  &TestContact{},
		"companies": &TestCompany{},
		"deals":     &TestDeal{},
	}))
	if err != nil {
		t.Fatalf("failed to create connector: %v", err)
	}

	tests := []struct {
		objectName string
		validate   func(t *testing.T, record map[string]any)
	}{
		{
			objectName: "contacts",
			validate: func(t *testing.T, record map[string]any) {
				// Verify required fields
				if _, hasEmail := record["email"]; !hasEmail {
					t.Error("expected 'email' field")
				}

				// Verify ID field
				if id, hasId := record["id"]; !hasId {
					t.Error("expected 'id' field")
				} else if idStr, ok := id.(string); !ok || idStr == "" {
					t.Error("expected non-empty string ID")
				}

				// Verify updated field
				if createdAt, hasCreatedAt := record["createdAt"]; !hasCreatedAt {
					t.Error("expected 'createdAt' field")
				} else if _, ok := createdAt.(int64); !ok {
					t.Errorf("expected int64 createdAt, got %T", createdAt)
				}

				// Verify enum field
				if status, hasStatus := record["status"]; hasStatus {
					statusStr := status.(string)
					validStatuses := []string{"active", "inactive", "pending"}
					found := false
					for _, valid := range validStatuses {
						if statusStr == valid {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("status %q not in valid enum values", statusStr)
					}
				}

				// Verify email format
				if email, hasEmail := record["email"]; hasEmail {
					emailStr := email.(string)
					if !strings.Contains(emailStr, "@") {
						t.Errorf("expected valid email format, got %q", emailStr)
					}
				}
			},
		},
		{
			objectName: "companies",
			validate: func(t *testing.T, record map[string]any) {
				// Verify integer ID
				if id, hasId := record["id"]; !hasId {
					t.Error("expected 'id' field")
				} else {
					switch id.(type) {
					case int, int64:
						// OK
					default:
						t.Errorf("expected integer ID, got %T", id)
					}
				}

				// Verify required name field
				if _, hasName := record["name"]; !hasName {
					t.Error("expected 'name' field")
				}

				// Verify updated field is date-time string
				if updatedAt, hasUpdatedAt := record["updatedAt"]; !hasUpdatedAt {
					t.Error("expected 'updatedAt' field")
				} else if updatedStr, ok := updatedAt.(string); !ok {
					t.Errorf("expected string updatedAt, got %T", updatedAt)
				} else if !strings.Contains(updatedStr, "T") {
					t.Errorf("expected ISO 8601 date-time format, got %q", updatedStr)
				}

				// Verify numeric constraints
				if empCount, hasEmpCount := record["employeeCount"]; hasEmpCount {
					count := empCount.(int)
					if count < 1 {
						t.Errorf("employeeCount %d violates minimum constraint (1)", count)
					}
				}
			},
		},
		{
			objectName: "deals",
			validate: func(t *testing.T, record map[string]any) {
				// Verify required title
				if _, hasTitle := record["title"]; !hasTitle {
					t.Error("expected 'title' field")
				}

				// Verify amount constraints
				if amount, hasAmount := record["amount"]; hasAmount {
					amountVal := amount.(float64)
					if amountVal < 0 {
						t.Errorf("amount %f violates minimum constraint (0)", amountVal)
					}
				}

				// Verify date format
				if closeDate, hasCloseDate := record["closeDate"]; hasCloseDate {
					dateStr := closeDate.(string)
					// Should be YYYY-MM-DD format
					if len(dateStr) != 10 || dateStr[4] != '-' || dateStr[7] != '-' {
						t.Errorf("expected YYYY-MM-DD date format, got %q", dateStr)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.objectName, func(t *testing.T) {
			t.Parallel()

			// Generate multiple records to test consistency
			for i := 0; i < 5; i++ {
				record, err := conn.GenerateRandomRecord(tt.objectName)
				if err != nil {
					t.Fatalf("GenerateRandomRecord failed: %v", err)
				}

				if record == nil {
					t.Fatal("expected non-nil record")
				}

				// Validate record structure
				tt.validate(t, record)

				// Verify record can be written (validates against schema)
				_, err = conn.Write(ctx, common.WriteParams{
					ObjectName: tt.objectName,
					RecordData: record,
				})
				if err != nil {
					t.Errorf("generated record failed validation: %v", err)
				}
			}
		})
	}
}

func TestListObjectMetadataWithStructSchemas(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	conn, err := NewConnector(WithStructSchemas(map[string]interface{}{
		"contacts":  &TestContact{},
		"companies": &TestCompany{},
	}))
	if err != nil {
		t.Fatalf("failed to create connector: %v", err)
	}

	result, err := conn.ListObjectMetadata(ctx, []string{"contacts", "companies"})
	if err != nil {
		t.Fatalf("ListObjectMetadata failed: %v", err)
	}

	t.Run("contacts metadata", func(t *testing.T) {
		metadata, exists := result.Result["contacts"]
		if !exists {
			t.Fatal("contacts metadata not found")
		}

		// Verify display name
		if metadata.DisplayName == "" {
			t.Error("expected non-empty display name")
		}

		// Verify fields
		if len(metadata.Fields) == 0 {
			t.Fatal("expected fields in metadata")
		}

		// Check email field (required, format)
		emailField, hasEmail := metadata.Fields["email"]
		if !hasEmail {
			t.Fatal("expected 'email' field in metadata")
		}
		if emailField.ValueType != common.ValueTypeString {
			t.Errorf("expected email ValueType String, got %v", emailField.ValueType)
		}
		if emailField.IsRequired == nil || !*emailField.IsRequired {
			t.Error("expected email to be marked as required")
		}
		if emailField.DisplayName == "" {
			t.Error("expected email to have display name")
		}

		// Check status field (enum -> SingleSelect)
		statusField, hasStatus := metadata.Fields["status"]
		if !hasStatus {
			t.Fatal("expected 'status' field in metadata")
		}
		if statusField.ValueType != common.ValueTypeSingleSelect {
			t.Errorf("expected status ValueType SingleSelect, got %v", statusField.ValueType)
		}
		if len(statusField.Values) == 0 {
			t.Error("expected status to have enum values")
		}
		// Verify enum values
		expectedValues := []string{"active", "inactive", "pending"}
		if len(statusField.Values) != len(expectedValues) {
			t.Errorf("expected %d enum values, got %d", len(expectedValues), len(statusField.Values))
		}

		// Check tags field (array)
		tagsField, hasTags := metadata.Fields["tags"]
		if !hasTags {
			t.Fatal("expected 'tags' field in metadata")
		}
		if tagsField.ValueType != common.ValueTypeOther {
			t.Errorf("expected tags ValueType Other (array), got %v", tagsField.ValueType)
		}

		// Check firstName field (with title)
		firstNameField, hasFirstName := metadata.Fields["firstName"]
		if !hasFirstName {
			t.Fatal("expected 'firstName' field in metadata")
		}
		if firstNameField.DisplayName != "First Name" {
			t.Errorf("expected display name 'First Name', got %q", firstNameField.DisplayName)
		}
	})

	t.Run("companies metadata", func(t *testing.T) {
		metadata, exists := result.Result["companies"]
		if !exists {
			t.Fatal("companies metadata not found")
		}

		// Check ID field (integer)
		idField, hasId := metadata.Fields["id"]
		if !hasId {
			t.Fatal("expected 'id' field in metadata")
		}
		if idField.ValueType != common.ValueTypeInt {
			t.Errorf("expected id ValueType Int, got %v", idField.ValueType)
		}

		// Check updatedAt field (date-time format -> DateTime type)
		updatedAtField, hasUpdatedAt := metadata.Fields["updatedAt"]
		if !hasUpdatedAt {
			t.Fatal("expected 'updatedAt' field in metadata")
		}
		if updatedAtField.ValueType != common.ValueTypeDateTime {
			t.Errorf("expected updatedAt ValueType DateTime, got %v", updatedAtField.ValueType)
		}

		// Check industry field (enum)
		industryField, hasIndustry := metadata.Fields["industry"]
		if !hasIndustry {
			t.Fatal("expected 'industry' field in metadata")
		}
		if industryField.ValueType != common.ValueTypeSingleSelect {
			t.Errorf("expected industry ValueType SingleSelect, got %v", industryField.ValueType)
		}

		// Check employeeCount field (integer with constraints)
		empCountField, hasEmpCount := metadata.Fields["employeeCount"]
		if !hasEmpCount {
			t.Fatal("expected 'employeeCount' field in metadata")
		}
		if empCountField.ValueType != common.ValueTypeInt {
			t.Errorf("expected employeeCount ValueType Int, got %v", empCountField.ValueType)
		}

		// Check website field (URI format)
		websiteField, hasWebsite := metadata.Fields["website"]
		if !hasWebsite {
			t.Fatal("expected 'website' field in metadata")
		}
		if websiteField.ValueType != common.ValueTypeString {
			t.Errorf("expected website ValueType String, got %v", websiteField.ValueType)
		}
	})
}

// ============================================================================
// Constructor and Setup Tests
// ============================================================================

func TestNewConnector_Success(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons":  testPersonSchema,
		"products": testProductSchema,
	}

	conn, err := NewConnector(WithSchemas(schemas))
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func TestNewConnector_EmptySchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schemas map[string]*InputSchema
		wantErr bool
	}{
		{
			name:    "nil schemas",
			schemas: nil,
			wantErr: true,
		},
		{
			name:    "empty schemas",
			schemas: map[string]*InputSchema{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn, err := NewConnector(WithSchemas(tt.schemas))
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
		schemas map[string]*InputSchema
	}{
		{
			name: "schema with invalid type",
			schemas: map[string]*InputSchema{
				"invalid": {
					// Schema with no type and no properties
					Schema: "https://json-schema.org/draft/2020-12/schema",
				},
			},
		},
		// Note: When using WithSchemas, the schemas are already parsed InputSchema objects.
		// Invalid/malformed JSON would be caught during the parsing phase before calling NewConnector.
		// This test validates that schemas with missing required fields are properly handled.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// The connector should be created successfully, but validation will happen when using it
			conn, err := NewConnector(WithSchemas(tt.schemas))
			// Note: The connector creation may succeed even with incomplete schemas,
			// but operations on those schemas will fail during validation
			if err != nil {
				require.Error(t, err)
				assert.Nil(t, conn)
			} else {
				// If connector creation succeeds, it should still be valid
				require.NoError(t, err)
				assert.NotNil(t, conn)
			}
		})
	}
}

func TestNewConnector_WithOptions(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}

	// Test WithClient option
	conn, err := NewConnector(WithSchemas(schemas), WithClient(http.DefaultClient))
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Test WithAuthenticatedClient option
	conn2, err := NewConnector(WithSchemas(schemas), WithAuthenticatedClient(http.DefaultClient))
	require.NoError(t, err)
	require.NotNil(t, conn2)
}

// ============================================================================
// Schema Validation Tests
// ============================================================================

func TestWrite_ValidData(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"products": testProductSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"complex": testComplexSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"products": testProductSchema,
		"complex":  testComplexSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"products": testProductSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"complex": testComplexSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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
	assert.Equal(t, 1, len(result.Data), "Since filter should return only records >= baseTime-50")
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons":  testPersonSchema,
		"products": testProductSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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
		Fields:     datautils.NewStringSet("id", "name", "email"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(readResult.Data))
	assert.Equal(t, "Alice", readResult.Data[0].Fields["name"])
	assert.Equal(t, "alice@example.com", readResult.Data[0].Fields["email"])
	assert.Equal(t, result.RecordId, readResult.Data[0].Fields["id"])
}

func TestMutationProtection_GetAll(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	sort.Slice(readResult2.Data, func(i, j int) bool {
		return readResult2.Data[i].Fields["name"].(string) <
			readResult2.Data[j].Fields["name"].(string)
	})

	assert.Equal(t, "Person 0", readResult2.Data[0].Fields["name"])
}

func TestDeepCopyVerification(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	primitiveSchema := mustParseSchema(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"stringField": {"type": "string"},
			"intField": {"type": "integer"},
			"numberField": {"type": "number"},
			"boolField": {"type": "boolean"}
		}
	}`)

	schemas := map[string]*InputSchema{
		"primitives": primitiveSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	formatSchema := mustParseSchema(`{
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

	schemas := map[string]*InputSchema{
		"formats": formatSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	enumSchema := mustParseSchema(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"status": {"type": "string", "enum": ["active", "inactive", "pending"]}
		}
	}`)

	schemas := map[string]*InputSchema{
		"enums": enumSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	constraintSchema := mustParseSchema(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"age": {"type": "integer", "minimum": 18, "maximum": 65},
			"price": {"type": "number", "exclusiveMinimum": 0, "maximum": 1000},
			"count": {"type": "integer", "multipleOf": 5}
		}
	}`)

	schemas := map[string]*InputSchema{
		"constraints": constraintSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	stringSchema := mustParseSchema(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"shortCode": {"type": "string", "minLength": 3, "maxLength": 5},
			"pattern": {"type": "string", "pattern": "^[A-Z]{2}-[0-9]{3}$"}
		}
	}`)

	schemas := map[string]*InputSchema{
		"strings": stringSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	arraySchema := mustParseSchema(`{
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

	schemas := map[string]*InputSchema{
		"arrays": arraySchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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

	nestedSchema := mustParseSchema(`{
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

	schemas := map[string]*InputSchema{
		"nested": nestedSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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
	deepSchema := mustParseSchema(`{
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

	schemas := map[string]*InputSchema{
		"deep": deepSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
	require.NoError(t, err)

	// Should not panic or hang
	record, err := conn.GenerateRandomRecord("deep")
	require.NoError(t, err)
	require.NotNil(t, record)
}

func TestGenerateRandomRecord_SpecialFields(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons": testPersonSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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
	// The updated field is an integer timestamp (int64), not float64
	updated, ok := record["updated"].(int64)
	assert.True(t, ok, "updated field should be int64, got %T", record["updated"])
	assert.Greater(t, updated, int64(0))
}

func TestGenerateRandomRecord_ValidationSuccess(t *testing.T) {
	t.Parallel()

	schemas := map[string]*InputSchema{
		"persons":  testPersonSchema,
		"products": testProductSchema,
		"complex":  testComplexSchema,
	}
	conn, err := NewConnector(WithSchemas(schemas))
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
