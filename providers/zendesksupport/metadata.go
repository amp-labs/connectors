package zendesksupport

import (
	"context"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module.ID, objectNames)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	wg.Add(len(objectNames))

	for _, objectName := range objectNames {
		go c.enhanceMetadataCustomFields(ctx, &wg, metadataResult, objectName)
	}

	wg.Wait()

	return metadataResult, nil
}

func (c *Connector) enhanceMetadataCustomFields(
	ctx context.Context, wg *sync.WaitGroup, metadataResult *common.ListObjectMetadataResult, objectName string,
) {
	defer wg.Done()

	fields, err := c.requestCustomFields(ctx, objectName)
	if err != nil {
		metadataResult.Errors[objectName] = err

		return
	}

	// Attach fields to the object metadata.
	objectMetadata := metadataResult.Result[objectName]
	for _, field := range fields {
		objectMetadata.AddFieldMetadata(field.Title, common.FieldMetadata{
			DisplayName:  field.TitleInPortal,
			ValueType:    field.GetValueType(),
			ProviderType: field.Type,
			ReadOnly:     false,
			Values:       field.getValues(),
		})
	}
}
