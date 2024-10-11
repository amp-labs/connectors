package deep

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dpmetadata"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type StaticMetadata struct {
	holder dpmetadata.StaticMetadataHolder
}

func newStaticMetadata(holder *dpmetadata.StaticMetadataHolder) *StaticMetadata {
	return &StaticMetadata{holder: *holder}
}

func (c StaticMetadata) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return c.holder.Metadata.Select(objectNames)
}

func (c StaticMetadata) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "staticMetadata",
		Constructor: newStaticMetadata,
	}
}
