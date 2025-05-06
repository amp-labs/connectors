package keap

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module.ID, objectNames)
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
		objectMetadata := metadataResult.Result[objectName]
		for _, field := range fields {
			objectMetadata.FieldsMap[field.FieldName] = field.Label
		}
	}

	return metadataResult, nil
}
