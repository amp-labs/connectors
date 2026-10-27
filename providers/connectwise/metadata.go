package connectwise

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise/internal/customfields"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

const (
	virtualFieldContactEmail   = "AMPERSAND-defaultEmail"
	virtualFieldContactEmailId = "AMPERSAND-defaultEmailId"
	virtualFieldContactFax     = "AMPERSAND-defaultFax"
	virtualFieldContactFaxId   = "AMPERSAND-defaultFaxId"
	virtualFieldContactPhone   = "AMPERSAND-defaultPhone"
	virtualFieldContactPhoneId = "AMPERSAND-defaultPhoneId"

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

		if err = customfields.EnhanceObjectMetadata(ctx, c.Base, objectName, &objectMetadata); err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		if objectName == objectNameContacts {
			c.attachContactFields(&objectMetadata)
		}

		// Write the modified metadata back to the map
		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}

func (c *Connector) attachContactFields(objectMetadata *common.ObjectMetadata) {
	// Email fields.
	objectMetadata.AddFieldMetadata(virtualFieldContactEmail, common.FieldMetadata{
		DisplayName:  "Default Email",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})
	objectMetadata.AddFieldMetadata(virtualFieldContactEmailId, common.FieldMetadata{
		DisplayName:  "Default Email Type Id",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})

	// Fax fields.
	objectMetadata.AddFieldMetadata(virtualFieldContactFax, common.FieldMetadata{
		DisplayName:  "Default Fax",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})
	objectMetadata.AddFieldMetadata(virtualFieldContactFaxId, common.FieldMetadata{
		DisplayName:  "Default Fax Type Id",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})

	// Phone fields.
	objectMetadata.AddFieldMetadata(virtualFieldContactPhone, common.FieldMetadata{
		DisplayName:  "Default Phone",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})
	objectMetadata.AddFieldMetadata(virtualFieldContactPhoneId, common.FieldMetadata{
		DisplayName:  "Default Phone Type Id",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
		ReadOnly:     new(false),
		IsCustom:     new(false),
		IsRequired:   new(false),
	})
}
