package zendesksupport

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Schemas.Select(c.Module.ID, objectNames)
}
