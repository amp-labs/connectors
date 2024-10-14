package dpread

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

// NextPageBuilder produces URL as opaque string which later will be saved under common.ReadParams{}.NextPage.
// The callback should produce fully formed URL for the next page.
// By default, empty page token is returned indicating that there is no next page, so pagination is not supported.
type NextPageBuilder struct {
	// Build is a callback producing next page token.
	// URL is referring to current request and node is a response.
	// Sometimes response has all the needed information, but in some cases previous URL is altered to produce next.
	Build func(config common.ReadParams, url *urlbuilder.URL, node *ajson.Node) (string, error)
}

func (b NextPageBuilder) NextPage(config common.ReadParams, url *urlbuilder.URL, node *ajson.Node) (string, error) {
	if b.Build == nil {
		return "", nil
	}

	return b.Build(config, url, node)
}

func (b NextPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.NextPageBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(PaginationStep),
	}
}
