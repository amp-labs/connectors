package twenty

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// listObjectMetadata resolves the requested objects against the workspace's own schema, so custom objects and
// custom fields are described with the names the workspace gave them, and a field added after the connector
// was built is described on the next call. It does not list every object of the workspace.
// Inactive objects are reported as unsupported; inactive fields are left out, although Twenty still returns
// their column in records (seen live on v2.50.0), so Raw keeps them.
func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	objects, err := c.fetchObjects(ctx)
	if err != nil {
		return nil, err
	}

	result := common.NewListObjectMetadataResult()

	for _, objectName := range datautils.NewSetFromList(objectNames).List() {
		object, ok := activeObject(objects, objectName)
		if !ok {
			result.AppendError(objectName, common.ErrObjectNotSupported)

			continue
		}

		result.Result[objectName] = *object.objectMetadata()
	}

	return result, nil
}

func (o objectMetadata) objectMetadata() *common.ObjectMetadata {
	fields := make(common.FieldsMetadata, len(o.Fields))

	for _, field := range o.Fields {
		key, ok := field.recordKey()
		if !field.IsActive || !ok {
			continue
		}

		fields[key] = field.fieldMetadata()
	}

	displayName := o.LabelPlural
	if displayName == "" {
		displayName = o.NamePlural
	}

	return common.NewObjectMetadata(displayName, fields)
}

func (f fieldMetadata) fieldMetadata() common.FieldMetadata {
	displayName := f.Label
	if displayName == "" {
		displayName = f.Name
	}

	readOnly := f.readOnly()
	// A non-nullable field with no default must be given on create.
	required := !readOnly && !f.IsNullable && f.DefaultValue == nil

	metadata := common.FieldMetadata{
		DisplayName:  displayName,
		ValueType:    valueType(f.Type),
		ProviderType: f.Type,
		ReadOnly:     &readOnly,
		IsRequired:   &required,
		FieldId:      &f.ID,
	}

	if len(f.Options) != 0 && (metadata.ValueType == common.ValueTypeSingleSelect ||
		metadata.ValueType == common.ValueTypeMultiSelect) {
		metadata.Values = make([]common.FieldValue, 0, len(f.Options))

		for _, option := range f.Options {
			metadata.Values = append(metadata.Values, common.FieldValue{
				Value:        option.Value,
				DisplayValue: option.Label,
			})
		}
	}

	return metadata
}

// valueType maps Twenty field types to Ampersand value types. Composite types (ADDRESS, CURRENCY, EMAILS,
// FULL_NAME, LINKS, PHONES, ACTOR, RICH_TEXT, FILES) are objects in records, kept as they are, so they are
// "other"; so are types whose JSON shape was not checked live (NUMERIC, ARRAY, RAW_JSON).
// https://docs.twenty.com/user-guide/data-model/capabilities/fields
func valueType(fieldType string) common.ValueType {
	switch fieldType {
	case "TEXT", "UUID":
		return common.ValueTypeString
	case "BOOLEAN":
		return common.ValueTypeBoolean
	case "NUMBER", "POSITION":
		return common.ValueTypeFloat
	case "DATE":
		return common.ValueTypeDate
	case "DATE_TIME":
		return common.ValueTypeDateTime
	case "SELECT", "RATING":
		return common.ValueTypeSingleSelect
	case "MULTI_SELECT":
		return common.ValueTypeMultiSelect
	case "RELATION", "MORPH_RELATION":
		return common.ValueTypeReference
	default:
		return common.ValueTypeOther
	}
}
