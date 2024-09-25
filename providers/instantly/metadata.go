package instantly

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/instantly/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	return metadata.Schemas.Select(objectNames)
}
