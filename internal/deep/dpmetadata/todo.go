package dpmetadata

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type StaticMetadataHolder struct {
	// TODO scrapper package should be renamed
	Metadata *scrapper.ObjectMetadataResult
}

func (h StaticMetadataHolder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.StaticMetadataHolder,
		Constructor: handy.Returner(h),
	}
}
