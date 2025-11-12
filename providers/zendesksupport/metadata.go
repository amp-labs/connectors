package zendesksupport

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(common.ModuleRoot, objectNames)
	if err != nil {
		return nil, err
	}

	customObjectNames := objectsWithCustomFields[common.ModuleRoot].Intersection(datautils.NewSetFromList(objectNames))
	if len(customObjectNames) == 0 {
		return metadataResult, nil
	}

	// Custom fields are shared across each customObject, therefore make only 1 API call.
	// Each custom object will either get new fields or an Error field will be set.
	fields, err := c.fetchCustomTicketFields(ctx)

	for _, objectName := range customObjectNames {
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		// Attach fields to the object metadata.
		objectMetadata := metadataResult.Result[objectName]
		for _, field := range fields {
			objectMetadata.AddFieldMetadata(field.Title, common.FieldMetadata{
				DisplayName:  field.TitleInPortal,
				ValueType:    field.GetValueType(),
				ProviderType: field.Type,
				Values:       field.getValues(),
			})
		}
	}

	return metadataResult, nil
}
