package connectwise

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		fields, err := c.requestCustomFields(ctx, objectName)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		// Attach fields to the object metadata.
		// Get a reference to the metadata in the map so changes are persisted.
		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			// Object not found in result, skip it
			continue
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

		// Write the modified metadata back to the map
		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}
