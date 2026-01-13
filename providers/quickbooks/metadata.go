package quickbooks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

// ListObjectMetadata overrides the base SchemaProvider to enhance metadata with custom fields.
func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	baseResult, err := c.SchemaProvider.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		return nil, fmt.Errorf("failed to get base metadata: %w", err)
	}

	customFields, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		// Graceful degradation: return base metadata if GraphQL fails
		slog.Warn("Failed to fetch custom field definitions, continuing with base metadata only",
			"error", err)

		return baseResult, nil
	}

	for _, objectName := range objectNames {
		objectMetadataPtr := baseResult.GetObjectMetadata(objectName)
		if objectMetadataPtr == nil {
			continue
		}

		objectCustomFields := filterCustomFieldsByObject(customFields, objectName)

		for _, field := range objectCustomFields {
			objectMetadataPtr.AddFieldMetadata(field.Name, common.FieldMetadata{
				DisplayName:  field.Name,
				ValueType:    getFieldValueType(field),
				ProviderType: field.Type,
				Values:       nil,
			})
		}

		baseResult.Result[objectName] = *objectMetadataPtr
	}

	return baseResult, nil
}
