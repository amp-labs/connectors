package deepmock

import "errors"

var (
	// ErrSchemaNotFound is returned when the object schema doesn't exist.
	ErrSchemaNotFound = errors.New("schema not found")

	// ErrValidationFailed is returned when data doesn't match schema.
	ErrValidationFailed = errors.New("validation failed")

	// ErrMissingParam is returned when a required parameter is missing.
	ErrMissingParam = errors.New("missing required parameter")

	// ErrRecordNotFound is returned when a record ID doesn't exist.
	ErrRecordNotFound = errors.New("record not found")

	// ErrInvalidSchema is returned when schema parsing fails.
	ErrInvalidSchema = errors.New("invalid schema")

	// ErrSchemaConversion is returned when schema to metadata conversion fails.
	ErrSchemaConversion = errors.New("failed to convert schema to metadata")

	// ErrUniqueConstraint is returned when uniqueItems constraint cannot be satisfied.
	ErrUniqueConstraint = errors.New("uniqueItems constraint cannot be satisfied")

	// ErrUniqueValue is returned when a unique value cannot be generated.
	ErrUniqueValue = errors.New("failed to generate unique value")

	// ErrNoProperties is returned when schema has no properties.
	ErrNoProperties = errors.New("schema has no properties to inject extras into")

	// ErrEmptySchemas is returned when schemas map is nil or empty.
	ErrEmptySchemas = errors.New("schemas map cannot be nil or empty")

	// ErrNilStruct is returned when struct instance is nil.
	ErrNilStruct = errors.New("struct instance cannot be nil")

	// ErrInvalidType is returned when expected struct but got different type.
	ErrInvalidType = errors.New("expected struct or pointer to struct")

	// ErrSchemaGeneration is returned when schema generation fails.
	ErrSchemaGeneration = errors.New("failed to generate schema")

	// ErrMissingField is returned when schema is missing required field.
	ErrMissingField = errors.New("schema missing required field")

	// ErrInvalidRef is returned when $ref format is invalid.
	ErrInvalidRef = errors.New("invalid $ref format")

	// ErrMissingDef is returned when $ref points to non-existent definition.
	ErrMissingDef = errors.New("$ref points to non-existent definition")

	// ErrMissingDefs is returned when schema has $ref but no $defs.
	ErrMissingDefs = errors.New("schema has $ref but no $defs")
)
