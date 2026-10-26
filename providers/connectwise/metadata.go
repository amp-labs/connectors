package connectwise

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise/internal/contactmodel"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
	"github.com/amp-labs/connectors/providers/connectwise/internal/customfields"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		// Get a reference to the metadata in the map so changes are persisted.
		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			// Object not found in result, skip it
			continue
		}

		if err = customfields.EnhanceObjectMetadata(ctx, c.Base, objectName, &objectMetadata); err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		if objectName == core.ObjectNameContacts {
			contactmodel.EnhanceObjectMetadata(&objectMetadata)
		}

		// Write the modified metadata back to the map
		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}
