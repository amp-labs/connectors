// Package deepmock provides an in-memory mock connector with JSON Schema validation.
//
// # Overview
//
// The deepmock connector extends the basic mock connector with the following capabilities:
//   - JSON Schema Draft 2020-12 validation for all data operations
//   - Custom schema extensions for identifying ID and timestamp fields
//   - Thread-safe in-memory storage with deep copying to prevent mutations
//   - Random record generation based on schema definitions
//   - Full support for Read, Write, Delete, and ObjectMetadata operations
//
// # Differences from Mock Connector
//
// Unlike the standard mock connector, deepmock:
//   - Validates all data against JSON schemas before storage
//   - Automatically handles ID generation and timestamp management
//   - Provides deep copies of records to ensure data isolation
//   - Supports complex schema features including nested objects and arrays
//   - Generates realistic random data conforming to schema constraints
//
// # Schema Format
//
// Schemas are provided as JSON Schema Draft 2020-12 documents with custom extensions:
//
//	{
//	  "type": "object",
//	  "properties": {
//	    "id": {
//	      "type": "string",
//	      "format": "uuid",
//	      "x-amp-id-field": true
//	    },
//	    "name": {
//	      "type": "string",
//	      "minLength": 1
//	    },
//	    "email": {
//	      "type": "string",
//	      "format": "email"
//	    },
//	    "age": {
//	      "type": "integer",
//	      "minimum": 0,
//	      "maximum": 150
//	    },
//	    "updated_at": {
//	      "type": "integer",
//	      "x-amp-updated-field": true
//	    }
//	  },
//	  "required": ["name", "email"]
//	}
//
// # Custom Schema Extensions
//
// The deepmock connector recognizes two custom schema extensions:
//
//   - x-amp-id-field: Marks a field as the record identifier. This field will be
//     automatically generated if not provided during creation. Supports string (UUID)
//     and integer (timestamp-based) types.
//
//   - x-amp-updated-field: Marks a field as the last-updated timestamp. This field
//     will be automatically updated on create and update operations. Supports string
//     (RFC3339) and integer (Unix timestamp) types.
//
// # Thread Safety
//
// All storage operations are protected by RWMutex locks, making the connector safe
// for concurrent use. Records are deep-copied on storage and retrieval to prevent
// external mutations from affecting stored data.
//
// # Usage Example
//
//	// Define schemas
//	schemas := map[string][]byte{
//	    "user": []byte(`{
//	        "type": "object",
//	        "properties": {
//	            "id": {"type": "string", "x-amp-id-field": true},
//	            "name": {"type": "string"},
//	            "email": {"type": "string", "format": "email"}
//	        },
//	        "required": ["name", "email"]
//	    }`),
//	}
//
//	// Create connector with required schemas parameter
//	connector, err := deepmock.NewConnector(schemas)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create connector with custom HTTP client
//	customClient := &http.Client{Timeout: 10 * time.Second}
//	connector, err = deepmock.NewConnector(
//	    schemas,
//	    deepmock.WithClient(customClient),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write a record
//	result, err := connector.Write(ctx, common.WriteParams{
//	    ObjectName: "user",
//	    RecordData: map[string]any{
//	        "name": "John Doe",
//	        "email": "john@example.com",
//	    },
//	})
//
//	// Read records
//	readResult, err := connector.Read(ctx, common.ReadParams{
//	    ObjectName: "user",
//	})
//
//	// Generate random record
//	randomUser, err := connector.GenerateRandomRecord("user")
package deepmock
