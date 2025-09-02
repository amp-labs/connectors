package custom

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

const (
	// Data type LongTextArea must be at least 256 in length.
	longTextAreaLength     = 256
	fieldTypeText          = "Text"
	fieldTypeLongText      = "LongTextArea"
	defaultNumDisplayLines = 10
)

type UpsertMetadataPayload struct {
	XMLName xml.Name `xml:"upsertMetadata"`

	// List of fields to create
	FieldsMetadata []UpsertMetadataCustomField `xml:"metadata"`
}

func NewCustomFieldsPayload(params *common.UpsertMetadataParams) UpsertMetadataPayload {
	fields := make([]UpsertMetadataCustomField, 0)

	for objectName, fieldDefinitions := range params.Fields {
		for _, fieldDefinition := range fieldDefinitions {
			fields = append(fields, newUpsertMetadataCustomField(objectName, fieldDefinition))
		}
	}

	return UpsertMetadataPayload{
		FieldsMetadata: fields,
	}
}

const metadataTypeCustomField = "CustomField"

// UpsertMetadataCustomField fields can be found here:
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/customfield.htm
type UpsertMetadataCustomField struct {
	XMLName               xml.Name `xml:"metadata"`
	AttributeMetadataType string   `xml:"xsi:type,attr"`

	// Common properties.
	FullName    string `xml:"fullName"`
	Label       string `xml:"label"`
	Description string `xml:"description"`
	Type        string `xml:"type"`
	Required    bool   `xml:"required"`
	Unique      bool   `xml:"unique"`
	Indexed     bool   `xml:"indexed"`

	// Special properties.
	Length               *int    `xml:"length,omitempty"`
	DefaultValue         any     `xml:"defaultValue,omitempty"`
	Formula              *string `xml:"formula,omitempty"`
	ReferenceTo          *string `xml:"referenceTo,omitempty"`
	ReferenceTargetField *string `xml:"referenceTargetField,omitempty"`
	DeleteConstraint     *string `xml:"deleteConstraint,omitempty"`
	RelationshipName     *string `xml:"relationshipName,omitempty"`
	Precision            *int    `xml:"precision,omitempty"`
	Scale                *int    `xml:"scale,omitempty"`

	// VisibleLines must be specified for field type `LongTextArea`.
	VisibleLines *string `xml:"visibleLines,omitempty"`

	// ValueSet is used for the picklists.
	ValueSet *ValueSet `xml:"valueSet,omitempty"`
}

func newUpsertMetadataCustomField(
	objectName string, field common.FieldDefinition,
) UpsertMetadataCustomField {
	result := UpsertMetadataCustomField{
		AttributeMetadataType: metadataTypeCustomField,
		FullName:              fmt.Sprintf("%v.%v", objectName, field.FieldName),
		Label:                 field.DisplayName,
		Description:           field.Description,
		Required:              field.Required,
		Unique:                field.Unique,
		Indexed:               field.Indexed,
	}

	result = handleTypeRequirements(result, field)
	result = enhanceWithStringOptions(result, field)
	result = enhanceWithNumericOptions(result, field)
	result = enhanceWithAssociation(result, field)

	return result
}

// Each data type requires unique payload setup.
// Date, Checkbox are straightforward, but others rely on String/Numeric/Association options.
func handleTypeRequirements(
	field UpsertMetadataCustomField, definition common.FieldDefinition,
) UpsertMetadataCustomField {
	field.Type = matchFieldType(definition)

	if definition.ValueType == common.ValueTypeString {
		return handleTypeString(field, definition)
	}

	if definition.Association != nil {
		return handleTypeAssociation(field, definition)
	}

	if definition.ValueType.IsSelectionType() {
		return handleTypeSelections(field, definition)
	}

	return field
}

// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_field_types.htm
func matchFieldType(definition common.FieldDefinition) string {
	switch definition.ValueType {
	case common.FieldTypeString:
		// Handled separately because it could be either `Text` or `LongTextArea`.
		return fieldTypeText
	case common.FieldTypeBoolean:
		return "Checkbox"
	case common.FieldTypeDate:
		return "Date"
	case common.FieldTypeDateTime:
		return "DateTime"
	case common.FieldTypeSingleSelect:
		return "Picklist"
	case common.FieldTypeMultiSelect:
		return "MultiselectPicklist"
	case common.FieldTypeInt, common.FieldTypeFloat:
		return "Number"
	default:
		return fieldTypeText
	}
}

// Automatically picks the type of field based on the length of the string.
func handleTypeString(field UpsertMetadataCustomField, definition common.FieldDefinition) UpsertMetadataCustomField {
	if definition.StringOptions == nil {
		return field
	}

	if definition.StringOptions.Length == nil {
		return field
	}

	if *definition.StringOptions.Length < longTextAreaLength {
		field.Type = fieldTypeText

		return field
	}

	field.Type = fieldTypeLongText
	// Must specify 'visibleLines' for a CustomField of type LongTextArea
	visibleLines := valueOrDefault(definition.StringOptions.NumDisplayLines, defaultNumDisplayLines)
	field.VisibleLines = goutils.Pointer(strconv.Itoa(visibleLines))

	return field
}

// This is limited at the moment. There is no lookup object in Ampersand but if association options
// are supplied then we automatically imply the Lookup field type.
// Note that there are other types, ex: ExternalLookup, IndirectLookup.
func handleTypeAssociation(
	field UpsertMetadataCustomField, definition common.FieldDefinition,
) UpsertMetadataCustomField {
	field.Type = "Lookup"

	if definition.Association.TargetField != "" {
		field.Type = "IndirectLookup"
	}

	return field
}

// Prepares the payload data structure which holds the options for Multi or Single select fields.
func handleTypeSelections(
	field UpsertMetadataCustomField, definition common.FieldDefinition,
) UpsertMetadataCustomField {
	if definition.ValueType == common.ValueTypeMultiSelect {
		// Must specify 'visibleLines' for a CustomField of type MultiselectPicklist.
		visibleLines := valueOrDefault(definition.StringOptions.NumDisplayLines, defaultNumDisplayLines)
		field.VisibleLines = goutils.Pointer(strconv.Itoa(visibleLines))
	}

	if definition.StringOptions != nil {
		values := make([]Value, len(definition.StringOptions.Values))

		for index, value := range definition.StringOptions.Values {
			isDefault := false
			if definition.StringOptions.DefaultValue != nil {
				isDefault = value == *definition.StringOptions.DefaultValue
			}

			values[index] = Value{
				FullName: value,
				Default:  isDefault,
				Label:    value,
			}
		}

		field.ValueSet = &ValueSet{
			Restricted: definition.StringOptions.ValuesRestricted,
			ValueSetDefinition: ValueSetDefinition{
				Sorted: true,
				Values: values,
			},
		}
	}

	return field
}

func enhanceWithStringOptions(
	field UpsertMetadataCustomField, definition common.FieldDefinition,
) UpsertMetadataCustomField {
	if definition.StringOptions == nil {
		return field
	}

	field.Length = definition.StringOptions.Length

	// Picklists don't specify the default value here. It is done inside ValueSet struct.
	// Client should pass this when working with Text or LongTextArea.
	if !definition.ValueType.IsSelectionType() {
		field.DefaultValue = definition.StringOptions.DefaultValue
	}

	if definition.StringOptions.Pattern != "" {
		field.Formula = goutils.Pointer(definition.StringOptions.Pattern)
	}

	return field
}

func enhanceWithNumericOptions(
	field UpsertMetadataCustomField, definition common.FieldDefinition,
) UpsertMetadataCustomField {
	if definition.NumericOptions == nil {
		return field
	}

	field.Precision = definition.NumericOptions.Precision
	field.Scale = definition.NumericOptions.Scale
	field.DefaultValue = definition.NumericOptions.DefaultValue

	return field
}

func enhanceWithAssociation(
	field UpsertMetadataCustomField, definition common.FieldDefinition,
) UpsertMetadataCustomField {
	if definition.Association == nil {
		return field
	}

	field.ReferenceTo = goutils.Pointer(definition.Association.TargetObject)

	if definition.Association.TargetField != "" {
		field.ReferenceTargetField = goutils.Pointer(definition.Association.TargetField)
	}

	if field.ReferenceTargetField == nil {
		field.DeleteConstraint = goutils.Pointer(string(definition.Association.OnDelete))
	}

	if definition.Association.ReverseLookupFieldName != "" {
		field.RelationshipName = goutils.Pointer(definition.Association.ReverseLookupFieldName)
	}

	return field
}

type ValueSet struct {
	XMLName            xml.Name           `xml:"valueSet"`
	Restricted         bool               `xml:"restricted"`
	ValueSetDefinition ValueSetDefinition `xml:"valueSetDefinition"`
}

type ValueSetDefinition struct {
	Sorted bool    `xml:"sorted"`
	Values []Value `xml:"value"`
}

type Value struct {
	FullName string `xml:"fullName"`
	Default  bool   `xml:"default"`
	Label    string `xml:"label"`
}

func valueOrDefault(value *int, defaultValue int) int {
	if value == nil {
		return defaultValue
	}

	return *value
}
