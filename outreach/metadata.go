package outreach

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {
	// TODO: To be implemented
	// In here to satisfy the Connector interface
	return nil, common.ErrNotImplemented
}
