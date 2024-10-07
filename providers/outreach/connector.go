package outreach

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion = "api/v2"
)

type Connector struct {
	deep.Clients
	deep.EmptyCloser
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.Outreach, interpreter.ErrorHandler{}).Build(opts)
}

func (c *Connector) getApiURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
