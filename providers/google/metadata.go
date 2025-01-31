package google

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Schemas.Select(c.Module.ID, objectNames)
}
