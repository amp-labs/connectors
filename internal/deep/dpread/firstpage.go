package dpread

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// FirstPageBuilder alters URL to satisfy first page requirements.
// A callback is used for customization. If not specified the url is unchanged.
type FirstPageBuilder struct {
	Build func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error)
}

func (b FirstPageBuilder) FirstPage(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
	if b.Build == nil {
		return url, nil
	}

	return b.Build(config, url)
}

func (b FirstPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.PaginationStartBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(PaginationStart),
	}
}
