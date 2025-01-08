package components

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// MetadataStrategy describes an object's schema / metadata retrieval strategy. It can be done using a REST API,
// OpenAPI file, etc.
type MetadataStrategy interface {
	GetObjectMetadata(ctx context.Context, objects ...string) (*common.ListObjectMetadataResult, error)
	fmt.Stringer
}

func (c *ConnectorComponent) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	if c.MetadataStrategy == nil {
		return nil, common.ErrNotImplemented
	}

	if len(objects) == 0 {
		return nil, common.ErrMissingObjects
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return c.MetadataStrategy.GetObjectMetadata(ctx, objects...)
}
