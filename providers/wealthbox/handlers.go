package wealthbox

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/wealthbox/metadata"
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

		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			continue
		}

		for _, field := range fields {
			objectMetadata.AddFieldMetadata(field.Name, common.FieldMetadata{
				DisplayName:  field.Name,
				ValueType:    field.getValueType(),
				ProviderType: field.FieldType,
				Values:       field.getValues(),
			})
		}

		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}
