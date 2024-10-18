package dpmetadata

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// SchemaHolder holds static metadata which was loaded from file.
// This holder can provide its data to connector components.
type SchemaHolder struct {
	Metadata *scrapper.ObjectMetadataResult
}

func (h SchemaHolder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.StaticMetadataHolder,
		Constructor: handy.PtrReturner(h),
	}
}
