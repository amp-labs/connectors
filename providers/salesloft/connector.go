package salesloft

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
}

func constructor(
	clients *deep.Clients,
	closer *deep.EmptyCloser,
) *Connector {
	return &Connector{
		Clients:     *clients,
		EmptyCloser: *closer,
	}
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](constructor, providers.Salesloft, opts,
		errorHandler,
	)
}

// Connector components.
var errorHandler = interpreter.ErrorHandler{ //nolint:gochecknoglobals
	JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
