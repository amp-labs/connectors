package deepmock

import (
	"fmt"
	"hash"
	"sort"

	"github.com/kaptinlin/jsonschema"
)

// InputSchemaMap is a mapping of schema names or identifiers to their corresponding InputSchema definitions.
// This is commonly used to organize and reference multiple schemas within a single document, such as
// in the $defs field or when managing a collection of reusable schema components.
type InputSchemaMap map[string]*InputSchema

// UpdateHash implements the hashable interface for InputSchemaMap, enabling deterministic
// hash computation for schema maps.
//
// The method ensures deterministic hashing by:
//   - Sorting map keys alphabetically before processing (Go maps have random iteration order)
//   - Using the hashBuilder to consistently encode each key-value pair
//   - Recursively hashing nested InputSchema values
//
// This allows InputSchemaMap instances with identical content to produce identical hashes
// regardless of insertion order or Go's internal map ordering, making it suitable for
// cache keys and content-addressable storage.
func (m InputSchemaMap) UpdateHash(h hash.Hash) error {
	builder := &hashBuilder{h: h}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		builder.String(k)
		builder.Hashable(m[k])
	}

	return builder.Error()
}

// InputSchema represents a JSON Schema Draft 2020-12 compliant schema definition.
// It contains all the keywords defined in the JSON Schema specification for validating
// and describing JSON data structures. This structure is used to define the expected
// shape and constraints of input data for mock generation.
//
// The schema supports:
//   - Type validation (string, number, object, array, boolean, null)
//   - Subschema composition (allOf, anyOf, oneOf, not)
//   - Conditional logic (if/then/else)
//   - String, numeric, array, and object constraints
//   - Schema references and definitions
//   - Content validation and metadata
//
// For detailed information about each keyword, see the JSON Schema specification at:
// https://json-schema.org/draft/2020-12/json-schema-core
type InputSchema struct {
	// Core schema identification and metadata
	// Public identifier for the schema, used as a base URI for resolving references.
	ID string `json:"$id,omitempty"`
	// URI indicating the specification version the schema conforms to.
	Schema string `json:"$schema,omitempty"`
	// Semantic format hint for string data validation (e.g., "email", "date-time", "uuid", "uri").
	Format *string `json:"format,omitempty"`

	// Schema reference keywords for reusability and composition
	// See https://json-schema.org/draft/2020-12/json-schema-core#ref
	// URI reference to another schema definition, enabling schema reuse.
	Ref string `json:"$ref,omitempty"`
	// Dynamic reference that can be overridden by enclosing schemas using $dynamicAnchor.
	DynamicRef string `json:"$dynamicRef,omitempty"`
	// Plain-name fragment identifier for the schema, used as a target for $ref.
	Anchor string `json:"$anchor,omitempty"`
	// Dynamic anchor that can be referenced by $dynamicRef for runtime schema resolution.
	DynamicAnchor string `json:"$dynamicAnchor,omitempty"`
	// Container for reusable schema definitions referenced via $ref (e.g., "#/$defs/address").
	Defs map[string]*InputSchema `json:"$defs,omitempty"`

	// Runtime-resolved references (not serialized to JSON)
	// Internally resolved schema pointer for $ref, populated during schema processing.
	ResolvedRef *InputSchema `json:"-"`
	// Internally resolved schema pointer for $dynamicRef, populated during schema processing.
	ResolvedDynamicRef *InputSchema `json:"-"`

	// Boolean JSON Schemas for shorthand validation
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-boolean-json-schemas
	// When true, always validates; when false, never validates. Used for quick accept/reject logic.
	Boolean *bool `json:"-"`

	// Logical composition keywords for combining multiple schemas
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subsch
	// Data must validate against ALL schemas in the array (intersection/AND logic).
	AllOf []*InputSchema `json:"allOf,omitempty"`
	// Data must validate against AT LEAST ONE schema in the array (union/OR logic).
	AnyOf []*InputSchema `json:"anyOf,omitempty"`
	// Data must validate against EXACTLY ONE schema in the array (exclusive OR logic).
	OneOf []*InputSchema `json:"oneOf,omitempty"`
	// Data must NOT validate against this schema (negation logic).
	Not *InputSchema `json:"not,omitempty"`

	// Conditional schema application based on validation results
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subsche
	// Condition schema; if data validates, Then is applied, otherwise Else is applied.
	If *InputSchema `json:"if,omitempty"`
	// Schema applied when If validates successfully (only evaluated when If is present and true).
	Then *InputSchema `json:"then,omitempty"`
	// Schema applied when If fails validation (only evaluated when If is present and false).
	Else *InputSchema `json:"else,omitempty"`
	// Schemas applied when specific properties are present; key is property name, value is schema.
	DependentSchemas map[string]*InputSchema `json:"dependentSchemas,omitempty"`

	// Schema application for array items
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subschem
	// Ordered schemas for validating array items by position (e.g., [0]=string, [1]=number).
	PrefixItems []*InputSchema `json:"prefixItems,omitempty"`
	// Schema applied to all array items (or items after prefixItems if both are present).
	Items *InputSchema `json:"items,omitempty"`
	// Schema that at least one array item must validate against; use with minContains/maxContains.
	Contains *InputSchema `json:"contains,omitempty"`

	// Schema application for object properties
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subschemas
	// Map of property names to their schemas for known object properties.
	Properties InputSchemaMap `json:"properties,omitempty"`
	// Map of regex patterns to schemas for properties matching those patterns.
	PatternProperties InputSchemaMap `json:"patternProperties,omitempty"`
	// Schema for properties not matched by Properties or PatternProperties; false disallows extras.
	AdditionalProperties *InputSchema `json:"additionalProperties,omitempty"`
	// Schema that all property names in the object must validate against.
	PropertyNames *InputSchema `json:"propertyNames,omitempty"`

	// Generic type validation keywords
	// See https://json-schema.org/draft/2020-12/json-schema-validation#section-6.1
	// Expected data type(s): "string", "number", "integer", "object", "array", "boolean", "null".
	Type jsonschema.SchemaType `json:"type,omitempty"`
	// Exhaustive list of allowed values; data must exactly match one of these values.
	Enum []any `json:"enum,omitempty"`
	// Single constant value that the data must exactly match.
	Const *jsonschema.ConstValue `json:"const,omitempty"`

	// Numeric validation constraints for number and integer types
	// See https://json-schema.org/draft/2020-12/json-schema-validation#section-6.2
	// Number must be divisible by this value (must be > 0).
	MultipleOf *jsonschema.Rat `json:"multipleOf,omitempty"`
	// Number must be less than or equal to this value (inclusive upper bound).
	Maximum *jsonschema.Rat `json:"maximum,omitempty"`
	// Number must be strictly less than this value (exclusive upper bound).
	ExclusiveMaximum *jsonschema.Rat `json:"exclusiveMaximum,omitempty"`
	// Number must be greater than or equal to this value (inclusive lower bound).
	Minimum *jsonschema.Rat `json:"minimum,omitempty"`
	// Number must be strictly greater than this value (exclusive lower bound).
	ExclusiveMinimum *jsonschema.Rat `json:"exclusiveMinimum,omitempty"`

	// String validation constraints
	// See https://json-schema.org/draft/2020-12/json-schema-validation#section-6.3
	MaxLength *float64 `json:"maxLength,omitempty"` // Maximum character count (must be >= 0).
	MinLength *float64 `json:"minLength,omitempty"` // Minimum character count (must be >= 0).
	Pattern   *string  `json:"pattern,omitempty"`   // ECMA-262 regular expression that the string must match.

	// Array validation constraints
	// See https://json-schema.org/draft/2020-12/json-schema-validation#section-6.4
	MaxItems    *float64 `json:"maxItems,omitempty"`    // Maximum array length (must be >= 0).
	MinItems    *float64 `json:"minItems,omitempty"`    // Minimum array length (must be >= 0).
	UniqueItems *bool    `json:"uniqueItems,omitempty"` // When true, all items must be unique.
	// Maximum number of items that can match the Contains schema (requires Contains).
	MaxContains *float64 `json:"maxContains,omitempty"`
	// Minimum number of items that must match the Contains schema (requires Contains, default=1).
	MinContains *float64 `json:"minContains,omitempty"`

	// Advanced array validation for items not covered by prefixItems or items
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-unevaluateditems
	// Schema for array items not evaluated by Items or PrefixItems; false disallows extra items.
	UnevaluatedItems *InputSchema `json:"unevaluatedItems,omitempty"`

	// Object validation constraints
	// See https://json-schema.org/draft/2020-12/json-schema-validation#section-6.5
	MaxProperties *float64 `json:"maxProperties,omitempty"` // Maximum number of properties (must be >= 0).
	// Minimum number of properties required in the object (must be >= 0).
	MinProperties *float64 `json:"minProperties,omitempty"`
	// Array of property names that must be present in the object.
	Required []string `json:"required,omitempty"`
	// When key property is present, all properties in the value array become required.
	DependentRequired map[string][]string `json:"dependentRequired,omitempty"`

	// Advanced object validation for properties not covered by Properties or PatternProperties
	// See https://json-schema.org/draft/2020-12/json-schema-core#name-unevaluatedproperties
	// Schema for properties not evaluated by Properties/PatternProperties; false disallows extras.
	UnevaluatedProperties *InputSchema `json:"unevaluatedProperties,omitempty"`

	// Content encoding and media type annotations for string data
	// See https://json-schema.org/draft/2020-12/json-schema-validation#name-a-vocabulary-for-the-conten
	ContentEncoding  *string `json:"contentEncoding,omitempty"`  // Content encoding (e.g., "base64").
	ContentMediaType *string `json:"contentMediaType,omitempty"` // MIME type (e.g., "application/json").
	// Schema to validate the decoded content against (used with ContentEncoding/ContentMediaType).
	ContentSchema *InputSchema `json:"contentSchema,omitempty"`

	// Descriptive metadata and documentation annotations
	// See https://json-schema.org/draft/2020-12/json-schema-validation#name-a-vocabulary-for-basic-meta
	Title       *string `json:"title,omitempty"`       // Human-readable title or label for the schema.
	Description *string `json:"description,omitempty"` // Detailed explanation of the schema's purpose.
	// Default value to use when the property is not provided (informational, not enforced).
	Default any `json:"default,omitempty"`
	// When true, indicates this schema is deprecated and should not be used.
	Deprecated *bool `json:"deprecated,omitempty"`
	// When true, value is managed by the server and should not be modified by clients.
	ReadOnly *bool `json:"readOnly,omitempty"`
	// When true, value can be sent to the server but will never be returned (e.g., passwords).
	WriteOnly *bool `json:"writeOnly,omitempty"`
	// Array of example values that validate against this schema (for documentation).
	Examples []any `json:"examples,omitempty"`

	// Ampersand-specific extension fields (vendor extensions prefixed with x-)
	// These custom fields are used internally by Ampersand to identify and handle special fields
	// in provider data models that have specific semantic meanings.
	// When true, marks this field as a unique identifier for the resource (e.g., "id", "userId").
	XAmpIdField *bool `json:"x-amp-id-field,omitempty"`
	// When true, marks field as a timestamp indicating when resource was last modified.
	XAmpUpdatedField *bool `json:"x-amp-updated-field,omitempty"`
}

// UpdateHash implements the hashable interface for InputSchema, enabling deterministic
// hash computation for JSON schema definitions.
//
// This method produces a stable hash for the entire schema structure by:
//   - Processing all schema fields in a fixed order (matching struct field declaration order)
//   - Sorting all map keys (Defs, DependentSchemas, DependentRequired) alphabetically before hashing
//   - Including array indices when hashing ordered collections (AllOf, AnyOf, OneOf, etc.)
//   - Using type-safe encoding via hashBuilder for each field type
//   - Recursively hashing nested InputSchema structures
//
// The deterministic nature ensures that:
//   - Identical schemas always produce identical hashes regardless of field insertion order
//   - Schema changes are reliably detected through hash comparison
//   - Schemas can be used as cache keys or for content-addressable storage
//
// All schema keywords defined in JSON Schema Draft 2020-12 are included in the hash,
// along with Ampersand-specific extension fields (x-amp-*).
//
// Note: The Examples field is intentionally excluded from the hash computation as it would
// add unnecessary complexity (requiring deep comparison of arbitrary values) while providing
// no meaningful value for schema identity - examples are purely documentation and don't affect
// validation behavior.
//
//nolint:cyclop,funlen // Complex schema hashing requires processing all JSON Schema keywords sequentially
func (is *InputSchema) UpdateHash(h hash.Hash) error {
	builder := &hashBuilder{h: h}

	builder.String(is.ID)
	builder.String(is.Schema)
	builder.StringPtr(is.Format)
	builder.String(is.Ref)
	builder.String(is.DynamicRef)
	builder.String(is.Anchor)
	builder.String(is.DynamicAnchor)

	defKeys := make([]string, 0, len(is.Defs))
	for k := range is.Defs {
		defKeys = append(defKeys, k)
	}

	sort.Strings(defKeys)

	for _, k := range defKeys {
		builder.String(k)
		builder.Hashable(is.Defs[k])
	}

	builder.Hashable(is.ResolvedRef)
	builder.Hashable(is.ResolvedDynamicRef)
	builder.BoolPtr(is.Boolean)

	for index, schema := range is.AllOf {
		builder.Int64(int64(index))
		builder.Hashable(schema)
	}

	for index, schema := range is.AnyOf {
		builder.Int64(int64(index))
		builder.Hashable(schema)
	}

	for index, schema := range is.OneOf {
		builder.Int64(int64(index))
		builder.Hashable(schema)
	}

	builder.Hashable(is.Not)

	builder.Hashable(is.If)
	builder.Hashable(is.Then)
	builder.Hashable(is.Else)

	dependentSchemaKeys := make([]string, 0, len(is.DependentSchemas))
	for k := range is.DependentSchemas {
		dependentSchemaKeys = append(dependentSchemaKeys, k)
	}

	sort.Strings(dependentSchemaKeys)

	for _, k := range dependentSchemaKeys {
		builder.String(k)
		builder.Hashable(is.DependentSchemas[k])
	}

	for index, schema := range is.PrefixItems {
		builder.Int64(int64(index))
		builder.Hashable(schema)
	}

	builder.Hashable(is.Items)
	builder.Hashable(is.Contains)

	builder.Hashable(is.Properties)
	builder.Hashable(is.PatternProperties)
	builder.Hashable(is.AdditionalProperties)
	builder.Hashable(is.PropertyNames)

	for _, tpe := range is.Type {
		builder.String(tpe)
	}

	for index, value := range is.Enum {
		builder.Int64(int64(index))
		builder.String(fmt.Sprint(value))
	}

	if is.Const == nil {
		builder.Nil()
	} else {
		builder.NonNil()
		builder.Bool(is.Const.IsSet)
		builder.String(fmt.Sprint(is.Const.Value))
	}

	builder.Rat(is.MultipleOf)
	builder.Rat(is.Maximum)
	builder.Rat(is.ExclusiveMaximum)
	builder.Rat(is.Minimum)
	builder.Rat(is.ExclusiveMinimum)

	builder.Float64Ptr(is.MaxLength)
	builder.Float64Ptr(is.MinLength)
	builder.StringPtr(is.Pattern)

	builder.Float64Ptr(is.MaxItems)
	builder.Float64Ptr(is.MinItems)
	builder.BoolPtr(is.UniqueItems)
	builder.Float64Ptr(is.MaxContains)
	builder.Float64Ptr(is.MinContains)

	builder.Hashable(is.UnevaluatedItems)

	builder.Float64Ptr(is.MaxProperties)
	builder.Float64Ptr(is.MinProperties)

	for index, fld := range is.Required {
		builder.Int64(int64(index))
		builder.String(fld)
	}

	dependentRequiredKeys := make([]string, 0, len(is.DependentRequired))
	for k := range is.DependentRequired {
		dependentRequiredKeys = append(dependentRequiredKeys, k)
	}

	sort.Strings(dependentRequiredKeys)

	for _, k := range dependentRequiredKeys {
		builder.String(k)

		for index, value := range is.DependentRequired[k] {
			builder.Int64(int64(index))
			builder.String(value)
		}
	}

	builder.Hashable(is.UnevaluatedProperties)

	builder.StringPtr(is.ContentEncoding)
	builder.StringPtr(is.ContentMediaType)
	builder.Hashable(is.ContentSchema)

	builder.StringPtr(is.Title)
	builder.StringPtr(is.Description)

	if is.Default == nil {
		builder.Nil()
	} else {
		builder.NonNil()
		builder.String(fmt.Sprint(is.Default))
	}

	builder.BoolPtr(is.ReadOnly)
	builder.BoolPtr(is.WriteOnly)

	builder.BoolPtr(is.XAmpIdField)
	builder.BoolPtr(is.XAmpUpdatedField)

	return builder.Error()
}
