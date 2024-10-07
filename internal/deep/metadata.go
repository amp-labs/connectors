package deep

import (
	"context"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type StaticMetadata struct {
	// TODO scrapper package should be renamed
	static scrapper.ObjectMetadataResult
}

func NewStaticMetadata(static *scrapper.ObjectMetadataResult) StaticMetadata {
	return StaticMetadata{static: *static}
}

func (c *StaticMetadata) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return c.static.Select(objectNames)
}
