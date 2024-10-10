package deep

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dpmetadata"
)

type StaticMetadata struct {
	holder dpmetadata.StaticMetadataHolder
}

func NewStaticMetadata(holder *dpmetadata.StaticMetadataHolder) *StaticMetadata {
	return &StaticMetadata{
		holder: *holder,
	}
}

func (c *StaticMetadata) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return c.holder.Metadata.Select(objectNames)
}
