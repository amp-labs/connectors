package common

type UpsertMetadataAction string

const (
	// UpsertMetadataActionCreate indicates that the object/field was created.
	UpsertMetadataActionCreate UpsertMetadataAction = "create"
	// UpsertMetadataActionUpdate indicates that the object/field was updated.
	UpsertMetadataActionUpdate UpsertMetadataAction = "update"
	// UpsertMetadataActionNone indicates that the object/field was not changed.
	UpsertMetadataActionNone UpsertMetadataAction = "none"
)

// UpsertMetadataParams matches the schema.yaml CreateMetadata structure.
type UpsertMetadataParams struct {
	// Maps object names to field definitions.
	FieldsDefinitions map[string][]*FieldDefinition `json:"customFields"`
}

type FieldType string

const (
	FieldTypeString       FieldType = "string"
	FieldTypeBoolean      FieldType = "boolean"
	FieldTypeDate         FieldType = "date"
	FieldTypeDateTime     FieldType = "datetime"
	FieldTypeSingleSelect FieldType = "singleSelect"
	FieldTypeMultiSelect  FieldType = "multiSelect"
	FieldTypeInt          FieldType = "int"
	FieldTypeFloat        FieldType = "float"
)

func (ft FieldType) IsValid() bool {
	switch ft {
	case
		FieldTypeString,
		FieldTypeBoolean,
		FieldTypeDate,
		FieldTypeDateTime,
		FieldTypeSingleSelect,
		FieldTypeMultiSelect,
		FieldTypeInt,
		FieldTypeFloat:
		return true
	default:
		return false
	}
}

// FieldDefinition represents a field definition. Note that not all
// providers will support all fields. This is a best-effort attempt
// to create a common schema for custom fields across providers.
//
// In the event that a provider doesn't support a particular field,
// and assuming that the field has a value, then the value should be
// ignored, and a warning should be added to the UpsertMetadataResult.
type FieldDefinition struct {
	// FieldName is the short name of the field, e.g. "My_Custom_Field".
	FieldName string `json:"fieldName"`
	// DisplayName is the human-readable name of the field, e.g. "My Custom Field".
	DisplayName string `json:"displayName"`
	// Description is an optional description of the field.
	Description string `json:"description,omitempty"`
	// ValueType is the data type of the field.
	ValueType FieldType `json:"valueType"`
	// Length is the maximum length of the field (for string types).
	Length *int `json:"length,omitempty"`
	// Precision is the total number of digits (for decimal types).
	Precision *int `json:"precision,omitempty"`
	// Scale is the number of digits to the right of the decimal point (for decimal types).
	Scale *int `json:"scale,omitempty"`
	// Required indicates if the field is required.
	Required bool `json:"required,omitempty"`
	// Unique indicates if the field must be unique across all records.
	Unique bool `json:"unique,omitempty"`
	// Indexed indicates if the field should be indexed for faster search.
	Indexed bool `json:"indexed,omitempty"`
	// DefaultValue is the default value for the field (if any).
	DefaultValue string `json:"defaultValue,omitempty"`
	// Min is the minimum value for numeric fields (if any).
	Min *float64 `json:"min,omitempty"`
	// Max is the maximum value for numeric fields (if any).
	Max *float64 `json:"max,omitempty"`
	// Pattern is a regex pattern that the field value must match (if any).
	Pattern string `json:"pattern,omitempty"`
	// Values is a list of allowed values for enum fields (if any).
	Values []string `json:"enumValues,omitempty"`
	// ValuesRestricted indicates if the field value must be limited to what's in Values.
	ValuesRestricted bool `json:"enumRestricted,omitempty"`
	// Association defines association/relationship information for the field (if any).
	Association *Relationship `json:"relationship,omitempty"`
}

// Relationship is a supporting type from schema.yaml.
type Relationship struct {
	Kind         string              `json:"kind"`
	TargetObject string              `json:"targetObject"`
	TargetField  string              `json:"targetField,omitempty"`
	Cardinality  string              `json:"cardinality,omitempty"`
	OnDelete     string              `json:"onDelete,omitempty"`
	Required     bool                `json:"required,omitempty"`
	InverseName  string              `json:"inverseName,omitempty"`
	Labels       *RelationshipLabels `json:"labels,omitempty"`
}

// RelationshipLabels is a supporting type from schema.yaml.
type RelationshipLabels struct {
	Singular string `json:"singular"`
	Plural   string `json:"plural"`
}

// UpsertMetadataResult contains results for all created/updated objects and fields.
type UpsertMetadataResult struct {
	// Indicates if the upsert operation was successful.
	Success bool `json:"success"`

	// Maps object names to field upsert results.
	Fields map[string][]*FieldUpsertResult `json:"fields"`
}

// FieldUpsertResult is a supporting type from schema.yaml.
type FieldUpsertResult struct {
	// FieldName is the name of the field.
	FieldName string `json:"fieldName"`
	// Action indicates what action was taken (create, update, none).
	Action UpsertMetadataAction `json:"action"`
	// Metadata contains provider-specific metadata about the field (if any).
	Metadata map[string]any `json:"metadata,omitempty"`
	// Warnings contains any warnings that occurred during the upsert operation,
	// such as unsupported field attributes.
	Warnings []string `json:"warnings,omitempty"`
}
