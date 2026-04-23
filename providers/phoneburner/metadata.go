package phoneburner

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/amp-labs/connectors/common"
	pbmetadata "github.com/amp-labs/connectors/providers/phoneburner/metadata"
)

// ListObjectMetadata returns static schema fields plus member-defined contact custom fields
// from GET /rest/1/customfields. See https://www.phoneburner.com/developer/route_list#customfields
func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	result, err := pbmetadata.Schemas.Select(c.ProviderContext.Module(), objectNames)
	if err != nil {
		return nil, fmt.Errorf("failed to get base metadata: %w", err)
	}

	if !slices.Contains(objectNames, objectContacts) {
		return result, nil
	}

	objectMetadata, ok := result.Result[objectContacts]
	if !ok {
		return result, nil
	}

	definitions, err := c.fetchMemberCustomFieldDefinitions(ctx)
	if err != nil {
		slog.Warn("Failed to fetch PhoneBurner custom field definitions; returning base contact metadata only",
			"error", err)

		return result, nil
	}

	if objectMetadata.Fields == nil {
		objectMetadata.Fields = make(common.FieldsMetadata)
	}

	for _, def := range definitions {
		if def.DisplayName == "" {
			continue
		}

		key := customFieldMetadataKey(def.DisplayName)
		isCustom := true

		objectMetadata.Fields[key] = common.FieldMetadata{
			DisplayName:  def.DisplayName,
			ValueType:    memberCustomFieldTypeToValueType(def.TypeID),
			ProviderType: def.TypeName,
			IsCustom:     &isCustom,
		}
	}

	result.Result[objectContacts] = objectMetadata

	return result, nil
}
