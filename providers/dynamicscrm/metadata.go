package dynamicscrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// ListObjectMetadata returns list of fields that can be queried during Read operation.
// The searched object will be considered not found by returning ErrObjectNotFound error.
// This will happen if API calls for schema/attributes fail.
// In an unlikely event it may happen if MS server's response format would change.
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	// enforce string formatting, then do internal delegation
	return c.listObjectMetadata(ctx, naming.NewSingularStrings(objectNames))
}

// Internal delegate. EntityDefinitions API uses single object names.
func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames naming.SingularStrings,
) (*common.ListObjectMetadataResult, error) {
	// Collect display name for the Object and its fields (name+display)
	metadataResult := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		metadataObjectName := objectName.Plural().String()

		objectDisplayName, err := c.fetchObjectDisplayName(ctx, objectName)
		if err != nil {
			metadataResult.AppendError(metadataObjectName, err)
		}

		fields, err := c.fetchFieldsForObject(ctx, objectName)
		if err != nil {
			metadataResult.AppendError(metadataObjectName, err)
		}

		// Object names must be in plural.
		// Connectors Read/Write methods for MS Dynamics use plural form. Ex: Read('contacts')
		// The expectation is therefore to match, while schema API uses singular. Ex: `contact` schema
		metadataResult.Result[metadataObjectName] = *common.NewObjectMetadata(
			objectDisplayName, fields,
		)
	}

	return metadataResult, nil
}
