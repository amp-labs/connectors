package getresponse

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
)

// Objects that support account-level custom field definitions (GET /v3/custom-fields).
var objectsWithCustomFields = datautils.NewStringSet(objectContacts) // nolint:gochecknoglobals

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

	if !objectNamesIncludeCustomFields(objectNames) {
		return metadataResult, nil
	}

	custom, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		// Best-effort: return static metadata when custom-fields API fails.
		return metadataResult, nil //nolint:nilerr // intentional: do not fail ListObjectMetadata on enrichment errors
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

func objectNamesIncludeCustomFields(objectNames []string) bool {
	for _, objectName := range objectNames {
		if objectsWithCustomFields.Has(objectName) {
			return true
		}
	}

	return false
}
