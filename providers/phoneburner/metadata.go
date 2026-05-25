package phoneburner

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/amp-labs/connectors/common"
	pbmetadata "github.com/amp-labs/connectors/providers/phoneburner/metadata"
)

// ListObjectMetadata returns static schema fields from embedded OpenAPI metadata.
// Only contacts may be augmented with member-defined custom fields (see mergeContactsCustomFieldMetadata).
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

	// Contacts are the only object with member custom field definitions from the provider API.
	if slices.Contains(objectNames, objectContacts) {
		c.mergeContactsCustomFieldMetadata(ctx, result)
	}

	return result, nil
}

// mergeContactsCustomFieldMetadata augments result.Result[contacts] with fields from
// GET /rest/1/customfields. Other objects in result are untouched.
// See https://www.phoneburner.com/developer/route_list#customfields
func (c *Connector) mergeContactsCustomFieldMetadata(ctx context.Context, result *common.ListObjectMetadataResult) {
	objectMetadata, ok := result.Result[objectContacts]
	if !ok {
		return
	}

	definitions, err := c.fetchMemberCustomFieldDefinitions(ctx)
	if err != nil {
		slog.Warn("Failed to fetch PhoneBurner custom field definitions; returning base contact metadata only",
			"error", err)

		return
	}

	if objectMetadata.Fields == nil {
		objectMetadata.Fields = make(common.FieldsMetadata)
	}

	for _, def := range definitions {
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
}
