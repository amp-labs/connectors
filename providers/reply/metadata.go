package reply

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/reply/metadata"
)

// ListObjectMetadata describes object fields from the static OpenAPI-derived
// schemas. For contacts and contact-accounts, the objects carrying user-defined
// custom fields, the custom field definitions are fetched live and merged in as
// native fields, named by their human-readable title.
//
// A failure to fetch the definitions fails the whole call, so callers don't
// silently miss custom fields.
// https://docs.reply.io/api-reference/custom-fields/list-all-custom-fields
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := metadata.Schemas.Select(common.ModuleRoot, objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		if !usesCustomFields(objectName) {
			continue
		}

		objectMetadata := result.GetObjectMetadata(objectName)
		if objectMetadata == nil {
			continue
		}

		definitions, err := c.exposedCustomFields(ctx, objectName)
		if err != nil {
			return nil, err
		}

		for _, definition := range definitions {
			objectMetadata.AddFieldMetadata(definition.Title, common.FieldMetadata{
				DisplayName:  definition.Title,
				ValueType:    definition.valueType(),
				ProviderType: definition.FieldType,
				IsCustom:     new(true),
			})
		}

		result.Result[objectName] = *objectMetadata
	}

	return result, nil
}
