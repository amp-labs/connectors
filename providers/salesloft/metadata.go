package salesloft

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesloft/internal/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := metadata.Schemas.Select(c.Module.ID, objectNames)
	if err != nil {
		return nil, err
	}

	return c.attachCustomMetadata(ctx, objectNames, result)
}
