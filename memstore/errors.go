package memstore

// This file defines error values that can be returned by the memstore connector.
//
// These errors are organized into categories:
//   - Schema errors: Issues with JSON schema definitions and validation
//   - Storage errors: Issues with record storage and retrieval
//   - Subscription errors: Issues with event subscriptions and observers
//   - Data generation errors: Issues with random data generation
//
// Most errors are sentinel values that can be checked using errors.Is() for
// consistent error handling across the connector.

import "errors"

var (
	// Schema and Validation Errors
	// These errors occur during schema processing, validation, and metadata conversion.

	// ErrSchemaNotFound is returned when attempting to access a schema for an object
	// type that doesn't exist in the connector's schema registry.
	// Common causes: misspelled object name, schema not registered during connector creation.
	ErrSchemaNotFound = errors.New("schema not found")

	// ErrValidationFailed is returned when data doesn't match the JSON schema constraints.
	// This can occur during Write operations or random data generation when the data
	// violates schema rules (type mismatches, missing required fields, constraint violations).
	ErrValidationFailed = errors.New("validation failed")

	// ErrInvalidSchema is returned when a JSON schema cannot be parsed or is malformed.
	// This typically occurs during connector initialization when schemas are loaded.
	ErrInvalidSchema = errors.New("invalid schema")

	// ErrSchemaConversion is returned when converting a JSON schema to object metadata fails.
	// This can happen if the schema is missing required fields or has an unsupported structure.
	ErrSchemaConversion = errors.New("failed to convert schema to metadata")

	// ErrNoProperties is returned when attempting to inject extra properties into a schema
	// that has no properties defined. This can occur during schema enhancement operations.
	ErrNoProperties = errors.New("schema has no properties to inject extras into")

	// ErrEmptySchemas is returned when attempting to create a connector with a nil or
	// empty schemas map. At least one object schema is required.
	ErrEmptySchemas = errors.New("schemas map cannot be nil or empty")

	// ErrMissingField is returned when a schema is missing a required field definition.
	// This can occur when validating schema completeness or during metadata conversion.
	ErrMissingField = errors.New("schema missing required field")

	// Storage Errors
	// These errors occur during record storage and retrieval operations.

	// ErrMissingParam is returned when a required parameter is missing from an API call.
	// Common examples: missing record ID in Get/Delete, missing object name in operations.
	ErrMissingParam = errors.New("missing required parameter")

	// ErrRecordNotFound is returned when attempting to retrieve or delete a record
	// that doesn't exist in storage. Can be checked with errors.Is() to distinguish
	// from other storage errors.
	ErrRecordNotFound = errors.New("record not found")

	// Subscription Errors
	// These errors occur during subscription and observer operations.

	// ErrObserverNotFound is returned when attempting to unsubscribe a subscription
	// that doesn't exist or has already been removed from storage.
	// Can be checked with errors.Is() for graceful handling of already-removed subscriptions.
	ErrObserverNotFound = errors.New("observer not found")

	// Data Generation Errors
	// These errors occur during random data generation operations.

	// ErrUniqueConstraint is returned when the uniqueItems constraint on an array
	// cannot be satisfied. This happens when the schema requires more unique items
	// than can be generated (e.g., an enum with 3 values but minItems: 5).
	ErrUniqueConstraint = errors.New("uniqueItems constraint cannot be satisfied")

	// ErrUniqueValue is returned when failing to generate a unique value after
	// multiple attempts. This can occur when the value space is small or heavily
	// constrained (e.g., string with very restrictive pattern and maxLength).
	ErrUniqueValue = errors.New("failed to generate unique value")

	// Schema Reference Errors
	// These errors occur when resolving JSON Schema $ref references.

	// ErrInvalidRef is returned when a $ref value has an invalid format.
	// Expected format: "#/$defs/DefinitionName" for local references.
	ErrInvalidRef = errors.New("invalid $ref format")

	// ErrMissingDef is returned when a $ref points to a definition name that
	// doesn't exist in the schema's $defs section.
	ErrMissingDef = errors.New("$ref points to non-existent definition")

	// ErrMissingDefs is returned when a schema contains $ref references but
	// has no $defs section to resolve them against.
	ErrMissingDefs = errors.New("schema has $ref but no $defs")

	// Struct Schema Errors
	// These errors occur when deriving schemas from Go struct types.

	// ErrNilStruct is returned when attempting to derive a schema from a nil
	// struct instance. A valid struct instance is required for reflection.
	ErrNilStruct = errors.New("struct instance cannot be nil")

	// ErrInvalidType is returned when attempting to derive a schema from a value
	// that is not a struct or pointer to struct. Only struct types are supported.
	ErrInvalidType = errors.New("expected struct or pointer to struct")

	// ErrSchemaGeneration is returned when automatic schema generation from a
	// struct fails due to reflection errors or unsupported field types.
	ErrSchemaGeneration = errors.New("failed to generate schema")

	// Association Errors
	// These errors occur during association expansion and validation operations.

	// ErrInvalidAssociation is returned when association metadata is malformed or missing required fields.
	// This can occur when an association schema doesn't specify the required fields like
	// AssociationType or TargetObject.
	ErrInvalidAssociation = errors.New("invalid association definition")

	// ErrAssociationTargetNotFound is returned when an association references a target object
	// that doesn't exist in the schema registry. This can happen if the target object name
	// in the association metadata is misspelled or the object wasn't registered.
	ErrAssociationTargetNotFound = errors.New("association target object not found")

	// ErrInvalidForeignKey is returned when attempting to write a record with a foreign key
	// field that references a non-existent record in the target object. This ensures referential
	// integrity is maintained during Write operations.
	ErrInvalidForeignKey = errors.New("foreign key references non-existent record")

	// Subscription Parameter Errors
	// These errors occur when subscription parameters are invalid or missing.

	// ErrSubscriptionNil is returned when a subscription context is nil.
	// This can occur when attempting to register a subscription without proper initialization.
	ErrSubscriptionNil = errors.New("subscription is nil")

	// ErrSubscriptionExists is returned when attempting to add a subscription with an ID
	// that already exists in storage. Each subscription must have a unique ID.
	ErrSubscriptionExists = errors.New("subscription already exists")

	// ErrRequestNil is returned when a required request parameter is nil.
	// This can occur in Register, Subscribe, or other subscription operations.
	ErrRequestNil = errors.New("request parameter is nil")

	// ErrInvalidRequestType is returned when a request parameter cannot be cast to the
	// expected type. This indicates a type mismatch in the subscription parameters.
	ErrInvalidRequestType = errors.New("request has invalid type")

	// ErrResultNil is returned when a required result parameter is nil.
	// This can occur when processing subscription or registration results.
	ErrResultNil = errors.New("result is nil")

	// ErrInvalidResultType is returned when a result parameter cannot be cast to the
	// expected type. This indicates a type mismatch in the result handling.
	ErrInvalidResultType = errors.New("result has invalid type")

	// ErrRegistrationResultNil is returned when the RegistrationResult parameter is nil
	// in a subscription operation. Registration must complete before subscribing.
	ErrRegistrationResultNil = errors.New("registration result is nil")

	// ErrInvalidRegistrationStatus is returned when the registration status is not Success.
	// Subscriptions can only be created from successful registrations.
	ErrInvalidRegistrationStatus = errors.New("invalid registration status")

	// ErrSubscriptionEventsEmpty is returned when no subscription events are specified.
	// At least one event type must be provided when creating a subscription.
	ErrSubscriptionEventsEmpty = errors.New("subscription events cannot be empty")

	// ErrEventTypeEmpty is returned when processing a subscription event with no event type.
	// All subscription events must have a valid event type (create, update, delete).
	ErrEventTypeEmpty = errors.New("event type is empty")

	// ErrObjectNameEmpty is returned when processing a subscription event with no object name.
	// All subscription events must specify which object type they relate to.
	ErrObjectNameEmpty = errors.New("object name is empty")

	// ErrRecordIDEmpty is returned when processing a subscription event with no record ID.
	// All subscription events must specify which record was affected.
	ErrRecordIDEmpty = errors.New("record ID is empty")
)
