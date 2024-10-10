package deep

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type StaticMetadata struct {
	holder StaticMetadataHolder
}

func NewStaticMetadata(holder *StaticMetadataHolder) *StaticMetadata {
	return &StaticMetadata{
		holder: *holder,
	}
}

func (c *StaticMetadata) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return c.holder.Metadata.Select(objectNames)
}

type StaticMetadataHolder struct {
	// TODO scrapper package should be renamed
	Metadata *scrapper.ObjectMetadataResult
}

func (h StaticMetadataHolder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "staticMetadataHolder",
		Constructor: handy.Returner(h),
	}
}
