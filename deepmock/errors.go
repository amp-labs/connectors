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
)
