package outreach

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion = "api/v2"
)

type Connector struct {
	*dprequests.Clients
	*deep.EmptyCloser
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(clients *dprequests.Clients, closer *deep.EmptyCloser) *Connector {
		return &Connector{
			Clients:     clients,
			EmptyCloser: closer,
		}
	}

	return deep.Connector[Connector, parameters](constructor, providers.Outreach, opts)
}

func (c *Connector) getApiURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
