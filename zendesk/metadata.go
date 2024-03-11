package zendesk

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// TODO:: Implement me
// Temporarily added as empty func to satisfy interface method requirements

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return &common.ListObjectMetadataResult{}, nil
}
