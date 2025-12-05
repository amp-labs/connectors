// nolint:revive,godoclint
package common

import "errors"

// UpsertMetadataAction represents the action taken during an upsert operation.
type UpsertMetadataAction string

const (
	// UpsertMetadataActionCreate indicates that the object/field was created.
	UpsertMetadataActionCreate UpsertMetadataAction = "create"
	// UpsertMetadataActionUpdate indicates that the object/field was updated.
	UpsertMetadataActionUpdate UpsertMetadataAction = "update"
	// UpsertMetadataActionNone indicates that the object/field was not changed.
	UpsertMetadataActionNone UpsertMetadataAction = "none"
)

// IsValid checks if the UpsertMetadataAction is known.
func (a UpsertMetadataAction) IsValid() bool {
	switch a {
	case UpsertMetadataActionCreate,
		UpsertMetadataActionUpdate,
		UpsertMetadataActionNone:
		return true
	default:
		return false
	}
}

// AssociationCardinality represents the cardinality of an association/relationship.
type AssociationCardinality string

const (
	// associationCardinalityNotSet means the cardinality is not specified.
	associationCardinalityNotSet AssociationCardinality = ""

	// AssociationCardinalityManyToOne indicates that many records in the referencing object
	// can reference a single record in the target object.
	AssociationCardinalityManyToOne AssociationCardinality = "many-to-one"

	// AssociationCardinalityOneToOne indicates that each record in the referencing object
	// can reference only one record in the target object, and vice versa.
	AssociationCardinalityOneToOne AssociationCardinality = "one-to-one"
)

// IsValid checks if the AssociationCardinality is known.
func (ac AssociationCardinality) IsValid() bool {
	switch ac {
	case associationCardinalityNotSet,
		AssociationCardinalityManyToOne,
		AssociationCardinalityOneToOne:
		return true
	default:
		return false
	}
}

type AssociationOnDeleteAction string

const (
	// associationOnDeleteActionNotSet means the action is not specified.
	associationOnDeleteActionNotSet AssociationOnDeleteAction = ""

	// AssociationOnDeleteActionCascade means that when the parent record is deleted,
	// all child records referencing it are also deleted.
	AssociationOnDeleteActionCascade AssociationOnDeleteAction = "cascade"

	// AssociationOnDeleteActionRestrict means that the parent record cannot be deleted
	// if there are any child records referencing it.
	AssociationOnDeleteActionRestrict AssociationOnDeleteAction = "restrict"

	// AssociationOnDeleteActionSetNull means that when the parent record is deleted,
	// the foreign key field in the child records is set to NULL.
	AssociationOnDeleteActionSetNull AssociationOnDeleteAction = "setNull"
)

func (a AssociationOnDeleteAction) IsValid() bool {
	switch a {
	case associationOnDeleteActionNotSet,
		AssociationOnDeleteActionCascade,
		AssociationOnDeleteActionRestrict,
		AssociationOnDeleteActionSetNull:
		return true
	default:
		return false
	}
}

// UpsertMetadataParams represents parameters for upserting metadata.
type UpsertMetadataParams struct {
	// Maps object names to field definitions.
	Fields map[string][]FieldDefinition `json:"fields"`
}

var ErrFieldTypeUnknown = errors.New("unrecognized field type")

// FieldType represents the data type of a field.
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

// IsValid checks if the FieldType is known.
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

func (ft FieldType) IsSelectionType() bool {
	return ft == FieldTypeSingleSelect || ft == FieldTypeMultiSelect
}

// FieldDefinition represents a field definition. Note that not all
// providers will support all fields. This is a best-effort attempt
// to create a common schema for custom fields across providers.
//
// In the event that a provider doesn't support a particular field,
// and assuming that the field has a value, then the value should be
// ignored, and a warning should be added to the UpsertMetadataResult.
type FieldDefinition struct {
	// FieldName is the identifier of the field, e.g. "My_Custom_Field".
	FieldName string `json:"fieldName"`
	// DisplayName is the human-readable name of the field, e.g. "My Custom Field".
	DisplayName string `json:"displayName"`
	// Description is an optional description of the field.
	Description string `json:"description,omitempty"`
	// ValueType is the data type of the field.
	ValueType FieldType `json:"valueType"`
	// Required indicates if the field is required.
	Required bool `json:"required,omitempty"`
	// Unique indicates if the field must be unique across all records.
	Unique bool `json:"unique,omitempty"`
	// Indexed indicates if the field should be indexed for faster search.
	Indexed bool `json:"indexed,omitempty"`
	// StringOptions contains additional options for string fields (if any).
	StringOptions *StringFieldOptions `json:"stringOptions,omitempty"`
	// NumericOptions contains additional options for numeric fields (if any).
	NumericOptions *NumericFieldOptions `json:"numericOptions,omitempty"`
	// Association defines association/relationship information for the field (if any).
	Association *AssociationDefinition `json:"association,omitempty"`
}

// NumericFieldOptions contains additional options for numeric fields.
// Note that not all providers will support all options.
// This is a best-effort attempt to create a common schema for numeric
// field options across providers.
//
// In the event that a provider doesn't support a particular option,
// and assuming that the option has a value, then the value should be
// ignored, and a warning should be added to the UpsertMetadataResult.
type NumericFieldOptions struct {
	// Precision is the total number of digits (for decimal types).
	Precision *int `json:"precision,omitempty"`
	// Scale is the number of digits to the right of the decimal point (for decimal types).
	Scale *int `json:"scale,omitempty"`
	// Min is the minimum value for numeric fields (if any).
	Min *float64 `json:"min,omitempty"`
	// Max is the maximum value for numeric fields (if any).
	Max *float64 `json:"max,omitempty"`
	// DefaultValue is the default value for the field (if any).
	DefaultValue *float64 `json:"defaultValue,omitempty"`
}

// StringFieldOptions contains additional options for string fields.
// Note that not all providers will support all options.
// This is a best-effort attempt to create a common schema for string
// field options across providers.
//
// In the event that a provider doesn't support a particular option,
// and assuming that the option has a value, then the value should be
// ignored, and a warning should be added to the UpsertMetadataResult.
type StringFieldOptions struct {
	// Length is the maximum length of the string field.
	Length *int `json:"length,omitempty"`
	// Pattern is a regex pattern that the string field value must match (if any).
	Pattern string `json:"pattern,omitempty"`
	// Values is a list of allowed values for enum fields (if any).
	Values []string `json:"values,omitempty"`
	// ValuesRestricted indicates if the field value must be limited to what's in Values.
	ValuesRestricted bool `json:"valuesRestricted,omitempty"`
	// DefaultValue is the default value for the field (if any).
	DefaultValue *string `json:"defaultValue,omitempty"`
	// NumDisplayLines defines how many lines of text are shown in the UI.
	// If the text exceeds this number, it will be truncated.
	NumDisplayLines *int `json:"lines,omitempty"`
}

// AssociationDefinition defines relationship information for a field
// to another object. Note that not all providers will support all
// aspects of the association. This is a best-effort attempt to create
// a common schema for associations across providers.
//
// In the event that a provider doesn't support a particular aspect of
// the association, and assuming that the field has a value, then the
// value should be ignored, and a warning should be added to the
// UpsertMetadataResult.
type AssociationDefinition struct {
	// AssociationType is the high-level association variety (e.g., 'foreignKey', 'lookup', 'ref').
	// The provider determines the exact behavior.
	AssociationType string `json:"associationType"`
	// TargetObject is the name of the referenced/parent object.
	TargetObject string `json:"targetObject"`
	// TargetField is the name of the referenced field on the target object.
	// Defaults to the target's primary key when omitted.
	TargetField string `json:"targetField,omitempty"`
	// Association cardinality from the referencing field's perspective (e.g., 'many-to-one', 'one-to-one').
	Cardinality AssociationCardinality `json:"cardinality,omitempty"`
	// OnDelete defines the behavior upon foreign object deletion, where supported.
	// E.g., 'cascade', 'restrict', 'setNull'.
	OnDelete AssociationOnDeleteAction `json:"onDelete,omitempty"`
	// Required means that, if true, a referenced record must exist (i.e., NOT NULL foreign key).
	Required bool `json:"required,omitempty"`
	// ReverseLookupFieldName is an optional inverse relationship/property name exposed on the target object.
	ReverseLookupFieldName string `json:"reverseLookupFieldName,omitempty"`
	// Labels represents optional UI labels for the association
	Labels *AssociationLabels `json:"labels,omitempty"`
}

// AssociationLabels represents UI labels for an association.
type AssociationLabels struct {
	Singular string `json:"singular"`
	Plural   string `json:"plural"`
}

// UpsertMetadataResult contains results for all created/updated objects and fields.
type UpsertMetadataResult struct {
	// Indicates if the upsert operation was successful.
	Success bool `json:"success"`

	// Maps object name -> field name -> upsert result.
	Fields map[string]map[string]FieldUpsertResult `json:"fields"`
}

// FieldUpsertResult is the result of an upsert operation for a single field.
// It indicates what action was taken (create, update, none) and any
// provider-specific metadata or warnings.
type FieldUpsertResult struct {
	// FieldName is the name of the field.
	FieldName string `json:"fieldName"`
	// Action indicates what action was taken (create, update, none).
	Action UpsertMetadataAction `json:"action"`
	// Metadata contains provider-specific metadata about the field (if any).
	// Specific keys/values will vary by provider. Considered strictly informational.
	Metadata map[string]any `json:"metadata,omitempty"`
	// Warnings contains any warnings that occurred during the upsert operation,
	// such as unsupported field attributes.
	Warnings []string `json:"warnings,omitempty"`
}
