package connectwise

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

const (
	virtualFieldContactEmail        = "AMPERSAND-email"
	virtualFieldContactEmailDefault = virtualFieldContactEmail + "-default"
	virtualFieldContactFax          = "AMPERSAND-fax"
	virtualFieldContactFaxDefault   = virtualFieldContactFax + "-default"
	virtualFieldContactPhone        = "AMPERSAND-phone"
	virtualFieldContactPhoneDefault = virtualFieldContactPhone + "-default"

	objectNameContacts = "contacts"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		// Get a reference to the metadata in the map so changes are persisted.
		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			// Object not found in result, skip it
			continue
		}

		if err = c.attachCustomFields(ctx, objectName, &objectMetadata); err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		if objectName == objectNameContacts {
			if err = c.attachContactFields(ctx, &objectMetadata); err != nil {
				metadataResult.Errors[objectName] = err

				continue
			}
		}

		// Write the modified metadata back to the map
		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}

func (c *Connector) attachCustomFields(ctx context.Context,
	objectName string,
	objectMetadata *common.ObjectMetadata,
) error {
	fields, err := c.requestCustomFields(ctx, objectName)
	if err != nil {
		return err
	}

	for _, field := range fields {
		fieldMetadata := common.FieldMetadata{
			DisplayName:  field.Caption,
			ValueType:    field.getValueType(),
			ProviderType: field.getProviderType(),
			ReadOnly:     new(field.ReadOnlyFlag),
			IsCustom:     new(true),
			IsRequired:   new(field.RequiredFlag),
			Values:       field.getValues(),
		}

		objectMetadata.AddFieldMetadata(field.makeFieldName(), fieldMetadata)
	}

	return nil
}

func (c *Connector) attachContactFields(ctx context.Context, objectMetadata *common.ObjectMetadata) error {
	types, err := c.requestCommunicationTypes(ctx)
	if err != nil {
		return err
	}

	for _, item := range types {
		field := common.FieldMetadata{
			DisplayName:  item.Description,
			ValueType:    common.ValueTypeString,
			ProviderType: "string",
			ReadOnly:     new(false),
			IsCustom:     new(false),
			IsRequired:   new(false),
		}

		var prefix string
		if item.EmailFlag {
			prefix = virtualFieldContactEmail
		}

		if item.FaxFlag {
			prefix = virtualFieldContactFax
		}

		if item.PhoneFlag {
			prefix = virtualFieldContactPhone
		}

		objectMetadata.AddFieldMetadata(prefix+item.Id.String(), field)
	}

	objectMetadata.AddFieldMetadata(virtualFieldContactEmailDefault, common.FieldMetadata{
		DisplayName:  "Default communication item for Email",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})

	objectMetadata.AddFieldMetadata(virtualFieldContactFaxDefault, common.FieldMetadata{
		DisplayName:  "Default communication item for Fax",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})

	objectMetadata.AddFieldMetadata(virtualFieldContactPhoneDefault, common.FieldMetadata{
		DisplayName:  "Default communication item for Phone",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})

	return nil
}
