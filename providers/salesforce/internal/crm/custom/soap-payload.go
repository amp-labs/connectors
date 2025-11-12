package custom

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
)

const (
	// Data type LongTextArea must be at least 256 in length.
	longTextAreaLength     = 256
	fieldTypeText          = "Text"
	fieldTypeLongText      = "LongTextArea"
	defaultNumDisplayLines = 10
)

// UpsertMetadataPayload represents the request body for the Salesforce Metadata API
// `upsertMetadata` operation. It supports creating or updating various metadata specified via [M] generic.
type UpsertMetadataPayload[M any] struct {
	XMLName xml.Name `xml:"upsertMetadata"`

	// MetadataList contains the list of metadata components to create or update.
	MetadataList []M
}

type UpsertCustomFieldsPayload UpsertMetadataPayload[MetadataCustomField]

func NewCustomFieldsPayload(params *common.UpsertMetadataParams) (*UpsertCustomFieldsPayload, error) {
	fields := make([]MetadataCustomField, 0)

	for objectName, fieldDefinitions := range params.Fields {
		for _, fieldDefinition := range fieldDefinitions {
			field, err := newMetadataCustomField(objectName, fieldDefinition)
			if err != nil {
				return nil, err
			}

			fields = append(fields, *field)
		}
	}

	return &UpsertCustomFieldsPayload{
		MetadataList: fields,
	}, nil
}

func (p UpsertCustomFieldsPayload) getOptionalFields() FieldPermissions {
	fields := make(FieldPermissions)

	for _, meta := range p.MetadataList {
		if !meta.Required {
			name := meta.FullName
			fields[name] = FieldPermission{
				FullName: name,
				Readable: true,
				Editable: true,
			}
		}
	}

	return fields
}

const metadataTypeCustomField = "CustomField"

// MetadataCustomField fields can be found here:
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/customfield.htm
type MetadataCustomField struct {
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

func newMetadataCustomField(
	objectName string, definition common.FieldDefinition,
) (*MetadataCustomField, error) {
	fieldType, err := matchFieldType(definition)
	if err != nil {
		return nil, err
	}

	result := MetadataCustomField{
		AttributeMetadataType: metadataTypeCustomField,
		Type:                  fieldType,
		FullName:              fmt.Sprintf("%v.%v", objectName, definition.FieldName),
		Label:                 definition.DisplayName,
		Description:           definition.Description,
		Required:              definition.Required,
		Unique:                definition.Unique,
		Indexed:               definition.Indexed,
	}

	result = handleTypeRequirements(result, definition)
	result = enhanceWithStringOptions(result, definition)
	result = enhanceWithNumericOptions(result, definition)
	result = enhanceWithAssociation(result, definition)

	return &result, nil
}

// Each data type requires unique payload setup.
// Date, Checkbox are straightforward, but others rely on String/Numeric/Association options.
func handleTypeRequirements(
	field MetadataCustomField, definition common.FieldDefinition,
) MetadataCustomField {
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
func matchFieldType(definition common.FieldDefinition) (string, error) {
	switch definition.ValueType {
	case common.FieldTypeString:
		// Handled separately because it could be either `Text` or `LongTextArea`.
		return fieldTypeText, nil
	case common.FieldTypeBoolean:
		return "Checkbox", nil
	case common.FieldTypeDate:
		return "Date", nil
	case common.FieldTypeDateTime:
		return "DateTime", nil
	case common.FieldTypeSingleSelect:
		return "Picklist", nil
	case common.FieldTypeMultiSelect:
		return "MultiselectPicklist", nil
	case common.FieldTypeInt, common.FieldTypeFloat:
		return "Number", nil
	default:
		if definition.Association != nil {
			// Real type of the field is determined by the presence of Association struct.
			// This means that common.FieldDefinition.ValueType can be set to anything.
			return "", nil
		}

		return "", fmt.Errorf("%w, fieldName: %v", common.ErrFieldTypeUnknown, definition.FieldName)
	}
}

// Automatically picks the type of field based on the length of the string.
func handleTypeString(field MetadataCustomField, definition common.FieldDefinition) MetadataCustomField {
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
	field MetadataCustomField, definition common.FieldDefinition,
) MetadataCustomField {
	field.Type = "Lookup"

	if definition.Association.TargetField != "" {
		field.Type = "IndirectLookup"
	}

	return field
}

// Prepares the payload data structure which holds the options for Multi or Single select fields.
func handleTypeSelections(
	field MetadataCustomField, definition common.FieldDefinition,
) MetadataCustomField {
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
	field MetadataCustomField, definition common.FieldDefinition,
) MetadataCustomField {
	if definition.StringOptions == nil {
		return field
	}

	field.Length = definition.StringOptions.Length

	// Picklists don't specify the default value here. It is done inside ValueSet struct.
	// Client should pass this when working with Text or LongTextArea.
	if !definition.ValueType.IsSelectionType() {
		if definition.StringOptions.DefaultValue != nil {
			defaultValue := *definition.StringOptions.DefaultValue
			// For Text and LongTextArea fields, wrap the defaultValue in quotes so Salesforce
			// treats it as a string literal rather than a field reference in formulas.
			defaultValueFormat := "%q"
			if definition.ValueType == common.ValueTypeBoolean {
				// For booleans the value cannot be wrapped in quotes.
				defaultValueFormat = "%v"
			}

			field.DefaultValue = fmt.Sprintf(defaultValueFormat, defaultValue)
		}
	}

	if definition.StringOptions.Pattern != "" {
		field.Formula = goutils.Pointer(definition.StringOptions.Pattern)
	}

	return field
}

func enhanceWithNumericOptions(
	field MetadataCustomField, definition common.FieldDefinition,
) MetadataCustomField {
	if definition.NumericOptions == nil {
		return field
	}

	field.Precision = definition.NumericOptions.Precision
	field.Scale = definition.NumericOptions.Scale
	field.DefaultValue = definition.NumericOptions.DefaultValue

	return field
}

func enhanceWithAssociation(
	field MetadataCustomField, definition common.FieldDefinition,
) MetadataCustomField {
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

// ReadMetadataPayload used to read metadata objects.
type ReadMetadataPayload struct {
	XMLName xml.Name `xml:"readMetadata"`

	// Type is the object type. Ex: PermissionSet.
	Type string `xml:"type"`
	// FullNames is a name of object instance.
	FullNames string `xml:"fullNames"`
}

func NewReadPermissionSetPayload() *ReadMetadataPayload {
	return &ReadMetadataPayload{
		Type:      PermissionSetType,
		FullNames: DefaultPermissionSetName,
	}
}

type UpsertPermissionSetPayload UpsertMetadataPayload[MetadataPermissionSet]

func NewPermissionSetPayload(fieldPermissions FieldPermissions) *UpsertPermissionSetPayload {
	return &UpsertPermissionSetPayload{
		MetadataList: []MetadataPermissionSet{
			{
				AttributeMetadataType: PermissionSetType,
				FullName:              DefaultPermissionSetName,
				Label:                 DefaultPermissionSetLabel,
				Description:           DefaultPermissionSetDescription,
				Fields:                datautils.FromMap(fieldPermissions).Values(),
			},
		},
	}
}

// MetadataPermissionSet fields can be found here:
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_permissionset.htm
type MetadataPermissionSet struct {
	XMLName               xml.Name `xml:"metadata"`
	AttributeMetadataType string   `xml:"xsi:type,attr"`

	// Common properties.
	FullName    string `xml:"fullName"`
	Label       string `xml:"label"`
	Description string `xml:"description"`

	// Fields
	Fields []FieldPermission `xml:"fieldPermissions"`
}
