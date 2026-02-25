package workday

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

// ListObjectMetadata overrides the embedded SchemaProvider to augment static metadata
// with custom field definitions fetched from the Workday API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	baseResult, err := c.SchemaProvider.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		objectMetadata := baseResult.GetObjectMetadata(objectName)
		if objectMetadata == nil {
			continue
		}

		registry, err := c.fetchCustomFieldDefinitions(ctx, objectName)
		if err != nil {
			baseResult.AppendError(objectName, err)

			continue
		}

		for _, def := range registry {
			objectMetadata.AddFieldMetadata(def.Name(), common.FieldMetadata{
				DisplayName:  def.Descriptor,
				ValueType:    def.getValueType(),
				ProviderType: def.FieldType,
				IsCustom:     goutils.Pointer(true),
			})
		}

		baseResult.Result[objectName] = *objectMetadata
	}

	return baseResult, nil
}
