package kit

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/kit/metadata"
)

// ListObjectMetadata creates metadata of object via reading objects using Kit API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	return metadata.Schemas.Select(c.Module.ID, objectNames)
}
