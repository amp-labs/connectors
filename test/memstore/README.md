# MemStore Connector Examples

This directory contains practical examples demonstrating the MemStore connector's capabilities for testing and development purposes.

## Overview

The MemStore connector is a powerful testing utility that provides in-memory CRUD operations with full JSON Schema validation. It's designed for:
- **Testing integrations** without external dependencies
- **Rapid prototyping** of connector workflows
- **Demonstration** of Ampersand connector patterns
- **Development** of new features in isolation

## Directory Structure

```
test/memstore/
├── README.md              # This file
├── connector.go           # Helper with CRM-style sample schemas
├── write/
│   └── write.go          # Write operation examples
├── read/
│   └── read.go           # Read operation examples
└── metadata/
    └── metadata.go       # Metadata operation examples
```

## Sample Schemas

The examples use realistic CRM-style schemas defined in `connector.go`:

### Contacts
- **Fields**: id (string), email (required), firstName, lastName, phone, status (enum), createdAt (timestamp), tags (array)
- **Use case**: Contact management with validation
- **ID type**: String UUID
- **Timestamp field**: createdAt (integer)

### Companies
- **Fields**: id (integer), name (required), industry (enum), employeeCount (min: 1), website (URI), updatedAt (timestamp)
- **Use case**: Company/account tracking
- **ID type**: Integer
- **Timestamp field**: updatedAt (string, date-time format)

### Deals
- **Fields**: id (string), title (required), amount (min: 0), stage (enum), contactId, companyId, closeDate (date), lastModified (timestamp)
- **Use case**: Sales pipeline management with relationships
- **ID type**: String UUID
- **Timestamp field**: lastModified (integer)

## Running the Examples

### Write Operations

Demonstrates creating and updating records:

```bash
cd test/memstore/write
go run write.go
```

**What it demonstrates:**
- Creating contacts using `GenerateRandomRecord()`
- Creating companies with explicit field values
- Creating deals with relationships (foreign keys to contacts/companies)
- Updating existing contacts (status changes, tag additions)
- Bulk creation of multiple records

**Expected output:**
- Structured logs showing operation progress
- JSON-formatted created/updated records
- Record IDs for tracking

### Read Operations

Demonstrates querying and filtering records:

```bash
cd test/memstore/read
go run read.go
```

**What it demonstrates:**
- Reading all contacts with field filtering
- Pagination with `PageSize` and `NextPage` tokens
- Time-based filtering using `Since` parameter for incremental reads
- Reading from multiple object types (contacts, companies, deals)

**Expected output:**
- Sample records from each page
- Pagination statistics (page count, total records)
- Time-filtered results showing only recent modifications
- Summary of records across different object types

### Metadata Operations

Demonstrates schema introspection:

```bash
cd test/memstore/metadata
go run metadata.go
```

**What it demonstrates:**
- Listing metadata for all configured objects
- Inspecting field types, required status, and constraints
- Viewing enum values and format specifications
- Understanding how metadata reflects JSON Schema constraints

**Expected output:**
- Field lists for each object type
- Detailed field metadata (type, required, nullable, properties)
- Constraint information (enums, minimums, formats)
- Schema validation rules

## Modifying Schemas for Custom Testing

You can easily customize the schemas in `connector.go` to test different scenarios:

### Adding a New Object Type

```go
var taskSchemaJSON = []byte(`{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "object",
    "properties": {
        "id": {"type": "string", "x-amp-id-field": true},
        "title": {"type": "string"},
        "priority": {"type": "string", "enum": ["low", "medium", "high"]},
        "completed": {"type": "boolean"},
        "updatedAt": {"type": "integer", "x-amp-updated-field": true}
    },
    "required": ["title"]
}`)

// Add to GetMemStoreConnector:
schemas := map[string][]byte{
    "contacts":  contactSchemaJSON,
    "companies": companySchemaJSON,
    "deals":     dealSchemaJSON,
    "tasks":     taskSchemaJSON, // New object type
}
```

### Testing Different Constraints

**String patterns:**
```json
{
    "phoneNumber": {
        "type": "string",
        "pattern": "^\\+?[1-9]\\d{1,14}$"
    }
}
```

**Numeric ranges:**
```json
{
    "age": {
        "type": "integer",
        "minimum": 18,
        "maximum": 65,
        "exclusiveMaximum": true
    }
}
```

**Array constraints:**
```json
{
    "skills": {
        "type": "array",
        "items": {"type": "string"},
        "minItems": 1,
        "maxItems": 10,
        "uniqueItems": true
    }
}
```

**Nested objects:**
```json
{
    "address": {
        "type": "object",
        "properties": {
            "street": {"type": "string"},
            "city": {"type": "string"},
            "zipCode": {"type": "string", "pattern": "^\\d{5}$"}
        },
        "required": ["city"]
    }
}
```

### Testing Format Validation

Supported formats in `GenerateRandomRecord()`:
- `email` - Generates valid email addresses
- `uuid` - Generates UUID strings
- `date` - Generates YYYY-MM-DD dates
- `date-time` - Generates ISO 8601 timestamps
- `phone` - Generates phone numbers
- `uri` / `url` - Generates HTTP URLs

Example:
```json
{
    "properties": {
        "userEmail": {"type": "string", "format": "email"},
        "userId": {"type": "string", "format": "uuid"},
        "birthDate": {"type": "string", "format": "date"},
        "createdAt": {"type": "string", "format": "date-time"}
    }
}
```

## Key Features Demonstrated

### 1. Thread-Safe Operations
All examples can be run concurrently. The MemStore connector uses internal locking to ensure data consistency.

### 2. Schema Validation
Every write operation validates data against the JSON Schema, ensuring:
- Required fields are present
- Field types match schema definitions
- Enum values are from the allowed list
- Numeric values are within min/max bounds
- String patterns match regex constraints
- Array items meet uniqueness/length requirements

### 3. Data Isolation
The connector creates deep copies of all data to prevent mutation:
- Reading a record and modifying it doesn't affect storage
- Passing a map to Write and modifying it afterward doesn't affect storage
- Complete isolation between test runs

### 4. Random Data Generation
`GenerateRandomRecord()` creates realistic test data that:
- Conforms to all schema constraints
- Uses appropriate formats (emails, UUIDs, dates)
- Respects numeric ranges and string patterns
- Generates valid enum values
- Creates nested objects and arrays

### 5. Pagination Support
Read operations support standard pagination:
- `PageSize` controls records per page
- `NextPage` token for continuation
- `Done` flag indicates completion

### 6. Time-Based Filtering
Incremental reads using timestamp fields:
- `Since` parameter for changes after a timestamp
- `Until` parameter for changes before a timestamp
- Automatic detection of `x-amp-updated-field`

## Use Cases

### Integration Testing
```go
func TestSalesforceSync(t *testing.T) {
    conn := memstoretest.GetMemStoreConnector(ctx)

    // Create test data
    contact := conn.GenerateRandomRecord("contacts")
    result, err := conn.Write(ctx, common.WriteParams{
        ObjectName: "contacts",
        RecordData: contact,
    })

    // Test your sync logic
    // ...
}
```

### Development Workflow
Use MemStore to develop connector features without external dependencies:
1. Define schema matching target provider
2. Generate test data with `GenerateRandomRecord()`
3. Implement and test your logic
4. Switch to real connector when ready

### Schema Validation Testing
Test edge cases and constraint handling:
```go
// Test required field validation
invalidData := map[string]any{
    "firstName": "John",
    // Missing required "email" field
}
_, err := conn.Write(ctx, common.WriteParams{
    ObjectName: "contacts",
    RecordData: invalidData,
})
// err should not be nil
```

## Tips and Best Practices

1. **Start Simple**: Begin with basic CRUD operations before testing complex scenarios
2. **Use GenerateRandomRecord**: Let the connector create valid test data for you
3. **Check Errors**: Always validate that operations succeed/fail as expected
4. **Inspect Metadata**: Use metadata operations to understand available fields
5. **Test Constraints**: Verify that schema validation catches invalid data
6. **Clean Data Between Tests**: Each test should start with a fresh connector instance
7. **Use Realistic Schemas**: Model your schemas after real provider APIs for better testing

## Further Reading

- **JSON Schema Documentation**: https://json-schema.org/draft/2020-12/json-schema-validation.html
- **Ampersand Connector Interface**: See `common/ReadConnector`, `common/WriteConnector`
- **MemStore Implementation**: See `memstore/connector.go` for full implementation details

## Support

For questions or issues:
- Review the unit tests in `memstore/connector_test.go` for comprehensive examples
- Check the MemStore connector implementation in `memstore/`
- Refer to other connector test examples in `test/`
