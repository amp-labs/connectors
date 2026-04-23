package getresponse

import (
	"context"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
)

// listObjectMetadata returns static OpenAPI metadata and merges per-account
// custom field definitions into contacts, similar to the Sellsy connector.
// If the custom-fields API request fails, static metadata is still returned
// (best-effort) so read flows work when metadata cannot be fully enriched.
func (c *Connector) listObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(objectNames, objectContacts) {
		return metadataResult, nil
	}

	custom, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		return metadataResult, nil
	}

	contactObject, ok := metadataResult.Result[objectContacts]
	if !ok {
		return metadataResult, nil
	}

	for i := range custom {
		cf := custom[i]
		contactObject.AddFieldMetadata(CustomFieldKey(cf.CustomFieldId), cf.fieldMetadata())
	}

	return metadataResult, nil
}
