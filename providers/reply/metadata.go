package reply

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/reply/metadata"
)

// ListObjectMetadata describes object fields from the static OpenAPI-derived
// schemas. For contacts — the only object carrying user-defined custom fields —
// the custom field definitions are fetched live and merged in as native fields,
// named by their human-readable title.
// https://docs.reply.io/api-reference/custom-fields/list-all-custom-fields
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(common.ModuleRoot, objectNames)
	if err != nil {
		return nil, err
	}

	objectMetadata, ok := metadataResult.Result[objectCustomFieldsOwner]
	if !ok {
		// Custom fields exist on contacts only; no API call needed otherwise.
		return metadataResult, nil
	}

	definitions, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	for _, definition := range definitions {
		objectMetadata.AddFieldMetadata(definition.Title, common.FieldMetadata{
			DisplayName:  definition.Title,
			ValueType:    definition.valueType(),
			ProviderType: definition.FieldType,
		})
	}

	return metadataResult, nil
}
