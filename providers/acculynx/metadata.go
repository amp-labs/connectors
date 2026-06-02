package acculynx

import (
	"context"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx/metadata"
)

// ListObjectMetadata shadows the embedded SchemaProvider's method so we can
// enrich the static schema for "contacts" and "jobs" with custom-field
// metadata sourced live from /company-settings/custom-fields. Other objects
// pass through unchanged.
//
// Strict on definitions-fetch failure (matches copper, sellsy, salesloft):
// the whole call aborts so callers don't silently miss custom fields.
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := metadata.Schemas.Select(c.ProviderContext.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	if !slices.ContainsFunc(objectNames, usesCustomFields) {
		return result, nil
	}

	defs, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	for _, name := range objectNames {
		entity, ok := customFieldEntityByObject[name]
		if !ok {
			continue
		}

		objectMetadata := result.GetObjectMetadata(name)
		if objectMetadata == nil {
			continue
		}

		for _, def := range defs[entity] {
			objectMetadata.AddFieldMetadata(def.fieldName(), common.FieldMetadata{
				DisplayName:  def.Label,
				ValueType:    def.valueType(),
				ProviderType: def.FieldType,
				Values:       def.getValues(),
				IsCustom:     new(true),
			})
		}

		result.Result[name] = *objectMetadata
	}

	return result, nil
}
